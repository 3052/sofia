package sofia

import (
   "crypto/aes"
   "crypto/cipher"
   "encoding/binary"
   "errors"
   "fmt"
)

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

// Put writes the header to the given byte slice.
func (h *BoxHeader) Put(b []byte) {
   binary.BigEndian.PutUint32(b[0:4], h.Size)
   copy(b[4:8], h.Type[:])
}

// parseContainer iterates over generic boxes within a byte slice.
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

// --- Decryption ---

// Decrypt decrypts a segment's mdat boxes in-place using the provided key.
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
      return fmt.Errorf("invalid key length")
   }
   block, err := aes.NewCipher(key)
   if err != nil {
      return err
   }
   for i := 0; i < len(segmentBoxes); i++ {
      if segmentBoxes[i].Moof != nil {
         moof := segmentBoxes[i].Moof
         if i+1 >= len(segmentBoxes) || segmentBoxes[i+1].Mdat == nil {
            return fmt.Errorf("malformed segment")
         }
         mdat := segmentBoxes[i+1].Mdat
         decryptFragment(moof, mdat.Payload, block)
         i++
      }
   }
   return nil
}

func decryptFragment(moof *MoofBox, mdatData []byte, block cipher.Block) {
   traf, ok := moof.Traf()
   if !ok {
      return
   }
   currentMdatOffset := 0
   tfhd, trun := traf.Tfhd(), traf.Trun()
   if tfhd == nil || trun == nil {
      return
   }
   senc, ok := traf.Senc()
   if !ok {
      return
   }
   for i, sampleInfo := range trun.Samples {
      sampleSize := sampleInfo.Size
      if sampleSize == 0 {
         sampleSize = tfhd.DefaultSampleSize
      }
      if currentMdatOffset+int(sampleSize) > len(mdatData) {
         return
      }
      sampleData := mdatData[currentMdatOffset : currentMdatOffset+int(sampleSize)]
      currentMdatOffset += int(sampleSize)
      iv := senc.Samples[i].IV
      if len(iv) == 8 {
         paddedIV := make([]byte, 16)
         copy(paddedIV, iv)
         iv = paddedIV
      }
      stream := cipher.NewCTR(block, iv)
      if len(senc.Samples[i].Subsamples) == 0 {
         stream.XORKeyStream(sampleData, sampleData)
      } else {
         sampleOffset := 0
         for _, sub := range senc.Samples[i].Subsamples {
            sampleOffset += int(sub.BytesOfClearData)
            if sub.BytesOfProtectedData > 0 {
               end := sampleOffset + int(sub.BytesOfProtectedData)
               chunk := sampleData[sampleOffset:end]
               stream.XORKeyStream(chunk, chunk)
               sampleOffset = end
            }
         }
      }
   }
}
