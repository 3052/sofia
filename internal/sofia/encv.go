package mp4

import "fmt"

// EncvChild holds either a parsed box or raw data for a child of an 'encv' box.
type EncvChild struct {
   Sinf *SinfBox
   Raw  []byte
}

// EncvBox represents the 'encv' box (Encrypted Video).
// It has a fixed-size header area before its children.
type EncvBox struct {
   Header      BoxHeader
   EntryHeader []byte // Stores the fixed-size part of the VisualSampleEntry
   Children    []EncvChild
}

const visualSampleEntrySize = 78 // 8 for SampleEntry, 70 for VisualSampleEntry

// ParseEncv parses the 'encv' box from a byte slice.
func ParseEncv(data []byte) (EncvBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EncvBox{}, err
   }
   var encv EncvBox
   encv.Header = header

   // The 'encv' box is a VisualSampleEntry, which has a 78-byte header area
   // before any child boxes start.
   payloadOffset := 8 // Start after the main box header
   if len(data) < payloadOffset+visualSampleEntrySize {
      // This box is too small to have children, so its entire payload is the header.
      encv.EntryHeader = data[payloadOffset:header.Size]
      return encv, nil
   }

   // Capture the fixed-size header part
   encv.EntryHeader = data[payloadOffset : payloadOffset+visualSampleEntrySize]

   // The rest of the data contains the child boxes.
   boxData := data[payloadOffset+visualSampleEntrySize : header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         // Not enough data for a full header, stop parsing children.
         break
      }

      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 {
         return EncvBox{}, fmt.Errorf("invalid box size %d in encv child", boxSize)
      }
      if offset+boxSize > len(boxData) {
         return EncvBox{}, fmt.Errorf("box size %d exceeds parent encv bounds", boxSize)
      }

      childData := boxData[offset : offset+boxSize]
      var child EncvChild

      switch string(h.Type[:]) {
      case "sinf":
         sinf, err := ParseSinf(childData)
         if err != nil {
            return EncvBox{}, err
         }
         child.Sinf = &sinf
      default:
         child.Raw = childData
      }
      encv.Children = append(encv.Children, child)
      offset += boxSize

      if h.Size == 0 {
         break
      }
   }
   return encv, nil
}

// Encode encodes the 'encv' box to a byte slice.
func (b *EncvBox) Encode() []byte {
   // First, encode all children boxes into a single byte slice.
   var childrenContent []byte
   for _, child := range b.Children {
      if child.Sinf != nil {
         childrenContent = append(childrenContent, child.Sinf.Encode()...)
      } else if child.Raw != nil {
         childrenContent = append(childrenContent, child.Raw...)
      }
   }

   // The full content is the fixed entry header followed by the children.
   content := append(b.EntryHeader, childrenContent...)

   b.Header.Size = uint32(8 + len(content))
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
