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
   case b.Mdat != nil:
      return b.Mdat.Encode()
   // Moof, Sidx, Pssh removed from Encode as they are read-only inputs
   default:
      return b.Raw
   }
}

func Parse(data []byte) ([]Box, error) {
   var boxes []Box
   offset := 0
   for offset < len(data) {
      if len(data)-offset < 8 {
         break
      }
      var header BoxHeader
      if err := header.Parse(data[offset:]); err != nil {
         return nil, fmt.Errorf("error reading header at offset %d: %w", offset, err)
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(data) - offset
      }
      if boxSize < 8 {
         return nil, fmt.Errorf("invalid box size %d at offset %d", boxSize, offset)
      }
      if offset+boxSize > len(data) {
         return nil, fmt.Errorf("box '%s' at offset %d with size %d exceeds file bounds", string(header.Type[:]), offset, boxSize)
      }
      boxData := data[offset : offset+boxSize]
      var currentBox Box
      switch string(header.Type[:]) {
      case "moov":
         var moov MoovBox
         if err := moov.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Moov = &moov
      case "moof":
         var moof MoofBox
         if err := moof.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Moof = &moof
      case "mdat":
         var mdat MdatBox
         if err := mdat.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Mdat = &mdat
      case "sidx":
         var sidx SidxBox
         if err := sidx.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Sidx = &sidx
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Pssh = &pssh
      default:
         currentBox.Raw = boxData
      }
      boxes = append(boxes, currentBox)
      offset += boxSize
   }
   return boxes, nil
}

// FindMoov finds the first MoovBox in a slice of generic boxes.
func FindMoov(boxes []Box) (*MoovBox, bool) {
   for _, box := range boxes {
      if box.Moov != nil {
         return box.Moov, true
      }
   }
   return nil, false
}

// AllMoof finds all MoofBoxes in a slice of generic boxes.
func AllMoof(boxes []Box) []*MoofBox {
   var moofs []*MoofBox
   for _, box := range boxes {
      if box.Moof != nil {
         moofs = append(moofs, box.Moof)
      }
   }
   return moofs
}

// FindSidx finds the first SidxBox in a slice of generic boxes.
func FindSidx(boxes []Box) (*SidxBox, bool) {
   for _, box := range boxes {
      if box.Sidx != nil {
         return box.Sidx, true
      }
   }
   return nil, false
}

// Decrypt decrypts a segment's mdat boxes in-place using the provided key.
func Decrypt(segmentBoxes []Box, key []byte) error {
   // ... (Existing decryption logic unchanged, can be kept for utility)
   // If you are only doing unfragmentation and not manual decryption via this function,
   // you can remove this entirely. Keeping it for now as it's separate from Encode logic.
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
