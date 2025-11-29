package sofia

import (
   "crypto/aes"
   "crypto/cipher"
   "encoding/binary"
   "errors"
   "fmt"
)

// Missing is an error type used when a required child box is not found.
type Missing string

func (e Missing) Error() string {
   return fmt.Sprintf("box '%s' not found", string(e))
}

type BoxHeader struct {
   Size uint32
   Type [4]byte
}

func (h *BoxHeader) Parse(data []byte) error {
   if len(data) < 8 {
      return errors.New("not enough data for box header")
   }
   h.Size = binary.BigEndian.Uint32(data[0:4])
   copy(h.Type[:], data[4:8])
   return nil
}

func (h *BoxHeader) Encode() []byte {
   buf := make([]byte, 8)
   binary.BigEndian.PutUint32(buf[0:4], h.Size)
   copy(buf[4:8], h.Type[:])
   return buf
}

// parseContainer iterates over generic boxes within a byte slice.
// onChild is called with the parsed header and the full box data (including header).
func parseContainer(data []byte, onChild func(BoxHeader, []byte) error) error {
   offset := 0
   for offset < len(data) {
      var h BoxHeader
      if err := h.Parse(data[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(data) - offset
      }
      if boxSize < 8 || offset+boxSize > len(data) {
         return errors.New("invalid child box size")
      }

      // Pass the full box (header + payload) to the callback
      if err := onChild(h, data[offset:offset+boxSize]); err != nil {
         return err
      }
      offset += boxSize
   }
   return nil
}

type Box struct {
   Moov *MoovBox
   Moof *MoofBox
   Mdat *MdatBox
   Sidx *SidxBox
   Pssh *PsshBox
   Raw  []byte
}

func (b *Box) Encode() []byte {
   switch {
   case b.Moov != nil:
      return b.Moov.Encode()
   default:
      return b.Raw
   }
}

func Parse(data []byte) ([]Box, error) {
   var boxes []Box
   err := parseContainer(data, func(h BoxHeader, boxData []byte) error {
      var currentBox Box
      switch string(h.Type[:]) {
      case "moov":
         var moov MoovBox
         if err := moov.Parse(boxData); err != nil {
            return err
         }
         currentBox.Moov = &moov
      case "moof":
         var moof MoofBox
         if err := moof.Parse(boxData); err != nil {
            return err
         }
         currentBox.Moof = &moof
      case "mdat":
         var mdat MdatBox
         if err := mdat.Parse(boxData); err != nil {
            return err
         }
         currentBox.Mdat = &mdat
      case "sidx":
         var sidx SidxBox
         if err := sidx.Parse(boxData); err != nil {
            return err
         }
         currentBox.Sidx = &sidx
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(boxData); err != nil {
            return err
         }
         currentBox.Pssh = &pssh
      default:
         currentBox.Raw = boxData
      }
      boxes = append(boxes, currentBox)
      return nil
   })
   return boxes, err
}

// --- Finders ---

func FindMoov(boxes []Box) (*MoovBox, bool) {
   for _, box := range boxes {
      if box.Moov != nil {
         return box.Moov, true
      }
   }
   return nil, false
}

func AllMoof(boxes []Box) []*MoofBox {
   var moofs []*MoofBox
   for _, box := range boxes {
      if box.Moof != nil {
         moofs = append(moofs, box.Moof)
      }
   }
   return moofs
}

func FindSidx(boxes []Box) (*SidxBox, bool) {
   for _, box := range boxes {
      if box.Sidx != nil {
         return box.Sidx, true
      }
   }
   return nil, false
}

func FindMoofPtr(boxes []Box) *MoofBox {
   for _, box := range boxes {
      if box.Moof != nil {
         return box.Moof
      }
   }
   return nil
}

func FindMdatPtr(boxes []Box) *MdatBox {
   for _, box := range boxes {
      if box.Mdat != nil {
         return box.Mdat
      }
   }
   return nil
}

// --- Helpers ---

func patchDuration(boxData []byte, newDuration uint64) error {
   if len(boxData) < 32 {
      return errors.New("box too short to patch duration")
   }
   version := boxData[8]
   if version == 1 {
      const durationOffset = 32
      if len(boxData) < durationOffset+8 {
         return errors.New("box too short for v1 duration")
      }
      binary.BigEndian.PutUint64(boxData[durationOffset:], newDuration)
   } else {
      const durationOffset = 24
      if len(boxData) < durationOffset+4 {
         return errors.New("box too short for v0 duration")
      }
      if newDuration > 0xFFFFFFFF {
         return errors.New("duration overflows 32-bit field")
      }
      binary.BigEndian.PutUint32(boxData[durationOffset:], uint32(newDuration))
   }
   return nil
}

// --- Decrypt ---

func Decrypt(segmentBoxes []Box, key []byte) error {
   var isEncrypted bool
   for _, moof := range AllMoof(segmentBoxes) {
      if traf, ok := moof.Traf(); ok {
         if _, ok := traf.Senc(); ok {
            isEncrypted = true
            break
         }
      }
   }
   if !isEncrypted {
      return nil
   }
   if len(key) != 16 {
      return fmt.Errorf("invalid key length: expected 16, got %d", len(key))
   }
   block, err := aes.NewCipher(key)
   if err != nil {
      return fmt.Errorf("AES cipher error: %w", err)
   }
   for i := 0; i < len(segmentBoxes); i++ {
      if segmentBoxes[i].Moof != nil {
         moof := segmentBoxes[i].Moof
         if i+1 >= len(segmentBoxes) || segmentBoxes[i+1].Mdat == nil {
            return fmt.Errorf("malformed segment: moof at index %d is not followed by an mdat", i)
         }
         mdat := segmentBoxes[i+1].Mdat
         // Decrypt modifies payload in place
         err := decryptFragment(moof, mdat.Payload, block)
         if err != nil {
            return fmt.Errorf("failed to process fragment at index %d: %w", i, err)
         }
         i++
      }
   }
   return nil
}

func decryptFragment(moof *MoofBox, mdatData []byte, block cipher.Block) error {
   traf, ok := moof.Traf()
   if !ok {
      return nil
   }
   currentMdatOffset := 0
   tfhd, trun := traf.Tfhd(), traf.Trun()
   if tfhd == nil || trun == nil {
      return errors.New("traf is missing required boxes: tfhd or trun")
   }
   senc, ok := traf.Senc()
   if !ok {
      return nil
   }
   if len(trun.Samples) != len(senc.Samples) {
      return errors.New("sample count mismatch between trun and senc")
   }
   for i, sampleInfo := range trun.Samples {
      sampleSize := sampleInfo.Size
      if sampleSize == 0 {
         sampleSize = tfhd.DefaultSampleSize
      }
      if sampleSize == 0 {
         return fmt.Errorf("sample %d has zero size", i)
      }
      if currentMdatOffset+int(sampleSize) > len(mdatData) {
         return fmt.Errorf("mdat buffer exhausted at sample %d", i)
      }
      sampleData := mdatData[currentMdatOffset : currentMdatOffset+int(sampleSize)]
      currentMdatOffset += int(sampleSize)
      iv := senc.Samples[i].IV
      if len(iv) == 8 {
         paddedIV := make([]byte, 16)
         copy(paddedIV, iv)
         iv = paddedIV
      } else if len(iv) != 16 {
         return fmt.Errorf("invalid IV length for sample %d: got %d, want 16", i, len(iv))
      }
      stream := cipher.NewCTR(block, iv)
      if len(senc.Samples[i].Subsamples) == 0 {
         stream.XORKeyStream(sampleData, sampleData)
      } else {
         sampleOffset := 0
         for _, sub := range senc.Samples[i].Subsamples {
            sampleOffset += int(sub.BytesOfClearData)
            if sub.BytesOfProtectedData > 0 {
               endOfProtected := sampleOffset + int(sub.BytesOfProtectedData)
               protectedPortion := sampleData[sampleOffset:endOfProtected]
               stream.XORKeyStream(protectedPortion, protectedPortion)
               sampleOffset = endOfProtected
            }
         }
      }
   }
   return nil
}
