package mp4

import (
   "encoding/binary"
   "errors"
   "fmt"
)

type BoxHeader struct {
   Size uint32
   Type [4]byte
}

// Read parses a box header from a byte slice into the BoxHeader struct.
// It returns the number of bytes read (8) or an error.
func (h *BoxHeader) Read(data []byte) (int, error) {
   if len(data) < 8 {
      return 0, errors.New("not enough data for box header")
   }
   h.Size = binary.BigEndian.Uint32(data[0:4])
   copy(h.Type[:], data[4:8])
   return 8, nil
}

func (h BoxHeader) Write(data []byte) int {
   binary.BigEndian.PutUint32(data[0:4], h.Size)
   copy(data[4:8], h.Type[:])
   return 8
}

// Box is a generic wrapper for any top-level MP4 box.
type Box struct {
   Moov *MoovBox
   Moof *MoofBox
   Mdat *MdatBox
   Sidx *SidxBox
   Pssh *PsshBox
   Raw  []byte
}

// Encode selects the correct encoder based on the top-level box type.
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

// ParseFile reads a byte slice and parses it into a slice of generic Box structs.
func ParseFile(data []byte) ([]Box, error) {
   var boxes []Box
   offset := 0
   for offset < len(data) {
      if len(data)-offset < 8 {
         break
      }
      var boxHeader BoxHeader
      _, err := boxHeader.Read(data[offset:])
      if err != nil {
         return nil, fmt.Errorf("error reading header at offset %d: %w", offset, err)
      }

      boxSize := int(boxHeader.Size)
      if boxSize == 0 {
         boxSize = len(data) - offset
      }
      if boxSize < 8 {
         return nil, fmt.Errorf("invalid box size %d at offset %d", boxSize, offset)
      }
      if offset+boxSize > len(data) {
         return nil, fmt.Errorf("box '%s' at offset %d with size %d exceeds file bounds", string(boxHeader.Type[:]), offset, boxSize)
      }

      boxData := data[offset : offset+boxSize]
      var currentBox Box

      switch string(boxHeader.Type[:]) {
      case "moov":
         var moovBox MoovBox
         if err := moovBox.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Moov = &moovBox
      case "moof":
         var moofBox MoofBox
         if err := moofBox.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Moof = &moofBox
      case "mdat":
         var mdatBox MdatBox
         if err := mdatBox.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Mdat = &mdatBox
      case "sidx":
         var sidxBox SidxBox
         if err := sidxBox.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Sidx = &sidxBox
      case "pssh":
         var psshBox PsshBox
         if err := psshBox.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Pssh = &psshBox
      default:
         currentBox.Raw = boxData
      }
      boxes = append(boxes, currentBox)
      offset += boxSize
   }
   return boxes, nil
}
