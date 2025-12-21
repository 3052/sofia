package sofia

import (
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

func FindSidx(boxes []Box) (*SidxBox, bool) {
   for _, box := range boxes {
      if box.Sidx != nil {
         return box.Sidx, true
      }
   }
   return nil, false
}
