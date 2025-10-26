package mp4

import "fmt"

// Box is a generic wrapper for any top-level MP4 box.
// It is now correctly simplified to only contain fields for boxes that can
// actually appear at the top level of a file segment.
type Box struct {
   // Pointers to parsed top-level container/data boxes.
   Moov *MoovBox
   Moof *MoofBox
   Mdat *MdatBox
   Sidx *SidxBox
   Pssh *PsshBox

   // Raw holds the data for any other top-level box type that we don't
   // parse, such as 'ftyp', 'styp', 'free', etc. This ensures a
   // byte-perfect round trip.
   Raw []byte
}

// Encode selects the correct encoder based on the top-level box type contained
// within the wrapper and returns the encoded byte slice.
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
      // If no specific parsed type is present, return the raw data.
      return b.Raw
   }
}

// ParseFile reads a byte slice and parses it into a slice of generic Box structs.
// This function correctly handles only top-level boxes.
func ParseFile(data []byte) ([]Box, error) {
   var boxes []Box
   offset := 0
   for offset < len(data) {
      if len(data)-offset < 8 {
         break // Not enough data for a full box header
      }
      h, _, err := ReadBoxHeader(data[offset:])
      if err != nil {
         return nil, fmt.Errorf("error reading header at offset %d: %w", offset, err)
      }

      boxSize := int(h.Size)
      // According to the spec, a size of 0 means the box extends to the end of the file.
      if boxSize == 0 {
         boxSize = len(data) - offset
      }

      if boxSize < 8 {
         return nil, fmt.Errorf("invalid box size %d at offset %d", boxSize, offset)
      }

      if offset+boxSize > len(data) {
         return nil, fmt.Errorf("box '%s' at offset %d with size %d exceeds file bounds", string(h.Type[:]), offset, boxSize)
      }

      boxData := data[offset : offset+boxSize]
      var currentBox Box

      // This switch ONLY handles boxes that can appear at the top level of an MP4 file.
      // The parsing of nested boxes (like 'trak', 'enca', etc.) is correctly handled
      // recursively by their respective parent parsers (e.g., ParseMoov).
      switch string(h.Type[:]) {
      case "moov":
         b, err := ParseMoov(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Moov = &b
      case "moof":
         b, err := ParseMoof(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Moof = &b
      case "mdat":
         b, err := ParseMdat(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Mdat = &b
      case "sidx":
         b, err := ParseSidx(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Sidx = &b
      case "pssh":
         b, err := ParsePssh(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Pssh = &b
      default:
         // For any other box type found at the top level (like 'ftyp', 'styp', 'free'),
         // we store its raw data to ensure a perfect round trip.
         currentBox.Raw = boxData
      }
      boxes = append(boxes, currentBox)
      offset += boxSize
   }
   return boxes, nil
}
