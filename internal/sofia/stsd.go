package mp4

import "encoding/binary"

// StsdChild holds either a parsed box or raw data for a sample entry in an 'stsd' box.
type StsdChild struct {
   Encv *EncvBox
   Enca *EncaBox
   Raw  []byte
}

// StsdBox represents the 'stsd' box (Sample Description Box).
type StsdBox struct {
   Header     BoxHeader
   Version    byte
   Flags      [3]byte
   EntryCount uint32
   Children   []StsdChild
}

// ParseStsd parses the 'stsd' box from a byte slice.
func ParseStsd(data []byte) (StsdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return StsdBox{}, err
   }
   var stsd StsdBox
   stsd.Header = header

   // stsd is a FullBox; it has version, flags, and an entry count.
   stsd.Version = data[8]
   copy(stsd.Flags[:], data[9:12])
   stsd.EntryCount = binary.BigEndian.Uint32(data[12:16])

   boxData := data[16:header.Size]
   offset := 0
   // Loop over the number of entries specified in the stsd header
   for i := uint32(0); i < stsd.EntryCount && offset < len(boxData); i++ {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return StsdBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
      var child StsdChild

      switch string(h.Type[:]) {
      case "encv":
         encv, err := ParseEncv(childData)
         if err != nil {
            return StsdBox{}, err
         }
         child.Encv = &encv
      case "enca":
         enca, err := ParseEnca(childData)
         if err != nil {
            return StsdBox{}, err
         }
         child.Enca = &enca
      default:
         // Any other sample entry type is stored as raw data.
         child.Raw = childData
      }
      stsd.Children = append(stsd.Children, child)
      offset += int(h.Size)
   }
   return stsd, nil
}

// Encode encodes the 'stsd' box to a byte slice.
func (b *StsdBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Encv != nil {
         content = append(content, child.Encv.Encode()...)
      } else if child.Enca != nil {
         content = append(content, child.Enca.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }

   // Total size is box header(8) + full box header(8) + children content length
   b.Header.Size = uint32(8 + 8 + len(content))
   encoded := make([]byte, b.Header.Size)

   // Write box header
   b.Header.Write(encoded)

   // Write version, flags, and entry count
   encoded[8] = b.Version
   copy(encoded[9:12], b.Flags[:])
   binary.BigEndian.PutUint32(encoded[12:16], b.EntryCount)

   // Write children content
   copy(encoded[16:], content)

   return encoded
}
