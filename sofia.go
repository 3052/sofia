package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
)

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
   case b.Moof != nil:
      return b.Moof.Encode()
   case b.Mdat != nil:
      return b.Mdat.Encode()
   case b.Sidx != nil:
      return b.Sidx.Encode()
   case b.Pssh != nil:
      return b.Pssh.Encode()
   default:
      return b.Raw
   }
}

func ParseFile(data []byte) ([]Box, error) {
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

func AllTrafs(boxes []Box) []*TrafBox {
   var trafs []*TrafBox
   for _, box := range boxes {
      if box.Moof != nil {
         for _, child := range box.Moof.Children {
            if child.Traf != nil {
               trafs = append(trafs, child.Traf)
            }
         }
      }
   }
   return trafs
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
