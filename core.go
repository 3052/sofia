package sofia

import (
   "encoding/binary"
   "errors"
)

// --- BoxHeader ---
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

func (h *BoxHeader) Put(buffer []byte) {
   binary.BigEndian.PutUint32(buffer[0:4], h.Size)
   copy(buffer[4:8], h.Type[:])
}

// --- parseContainer ---
func parseContainer(data []byte, onChild func(BoxHeader, []byte) error) error {
   offset := 0
   for offset < len(data) {
      var header BoxHeader
      if err := header.Parse(data[offset:]); err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(data) - offset
      }
      if boxSize < 8 || offset+boxSize > len(data) {
         return errors.New("invalid child box size")
      }
      if err := onChild(header, data[offset:offset+boxSize]); err != nil {
         return err
      }
      offset += boxSize
   }
   return nil
}

// --- Box ---
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
   err := parseContainer(data, func(header BoxHeader, boxData []byte) error {
      var currentBox Box
      switch string(header.Type[:]) {
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

func FindSidx(boxes []Box) (*SidxBox, bool) {
   for _, box := range boxes {
      if box.Sidx != nil {
         return box.Sidx, true
      }
   }
   return nil, false
}

// --- MDAT ---
type MdatBox struct {
   Header  BoxHeader
   Payload []byte
}

func (b *MdatBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.Payload = data[8:b.Header.Size]
   return nil
}

// --- SIDX ---
type SidxReference struct {
   ReferenceType      bool
   ReferencedSize     uint32
   SubsegmentDuration uint32
   StartsWithSAP      bool
   SAPType            uint8
   SAPDeltaTime       uint32
}

type SidxBox struct {
   Header                   BoxHeader
   Version                  byte
   Flags                    uint32
   ReferenceID              uint32
   Timescale                uint32
   EarliestPresentationTime uint64
   FirstOffset              uint64
   References               []SidxReference
}

func (b *SidxBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return errors.New("sidx box too short")
   }
   b.Version = data[8]
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   offset := 12
   if len(data) < offset+8 {
      return errors.New("sidx box too short")
   }
   b.ReferenceID = binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   if b.Version == 0 {
      if len(data) < offset+8 {
         return errors.New("sidx v0 box too short")
      }
      b.EarliestPresentationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.FirstOffset = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   } else {
      if len(data) < offset+16 {
         return errors.New("sidx v1 box too short")
      }
      b.EarliestPresentationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.FirstOffset = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   }
   if len(data) < offset+4 {
      return errors.New("sidx box too short for reference_count")
   }
   offset += 2 // reserved
   referenceCount := binary.BigEndian.Uint16(data[offset : offset+2])
   offset += 2
   if len(data)-offset < int(referenceCount)*12 {
      return errors.New("sidx box too short for declared references")
   }
   b.References = make([]SidxReference, referenceCount)
   for i := 0; i < int(referenceCount); i++ {
      val1 := binary.BigEndian.Uint32(data[offset : offset+4])
      b.References[i].ReferenceType = (val1 >> 31) == 1
      b.References[i].ReferencedSize = val1 & 0x7FFFFFFF
      offset += 4
      b.References[i].SubsegmentDuration = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      val2 := binary.BigEndian.Uint32(data[offset : offset+4])
      b.References[i].StartsWithSAP = (val2 >> 31) == 1
      b.References[i].SAPType = uint8((val2 >> 28) & 0x07)
      b.References[i].SAPDeltaTime = val2 & 0x0FFFFFFF
      offset += 4
   }
   return nil
}
