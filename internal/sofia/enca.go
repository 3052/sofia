package mp4

import "fmt"

// EncaChild holds either a parsed box or raw data for a child of an 'enca' box.
type EncaChild struct {
   Sinf *SinfBox
   Raw  []byte
}

// EncaBox represents the 'enca' box (Encrypted Audio).
// It has a fixed-size header area before its children.
type EncaBox struct {
   Header      BoxHeader
   EntryHeader []byte // Stores the fixed-size part of the AudioSampleEntry
   Children    []EncaChild
}

const audioSampleEntrySize = 28 // 8 bytes for SampleEntry, 20 for AudioSampleEntry

// ParseEnca parses the 'enca' box from a byte slice.
func ParseEnca(data []byte) (EncaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EncaBox{}, err
   }
   var enca EncaBox
   enca.Header = header

   // The 'enca' box is an AudioSampleEntry, which has a 28-byte header area
   // before any child boxes start.
   payloadOffset := 8 // Start after the main box header
   if len(data) < payloadOffset+audioSampleEntrySize {
      // This box is too small to have children, so its entire payload is the header.
      enca.EntryHeader = data[payloadOffset:header.Size]
      return enca, nil
   }

   // Capture the fixed-size header part
   enca.EntryHeader = data[payloadOffset : payloadOffset+audioSampleEntrySize]

   // The rest of the data contains the child boxes.
   boxData := data[payloadOffset+audioSampleEntrySize : header.Size]
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
         return EncaBox{}, fmt.Errorf("invalid box size %d in enca child", boxSize)
      }
      if offset+boxSize > len(boxData) {
         return EncaBox{}, fmt.Errorf("box size %d exceeds parent enca bounds", boxSize)
      }

      childData := boxData[offset : offset+boxSize]
      var child EncaChild

      switch string(h.Type[:]) {
      case "sinf":
         sinf, err := ParseSinf(childData)
         if err != nil {
            return EncaBox{}, err
         }
         child.Sinf = &sinf
      default:
         child.Raw = childData
      }
      enca.Children = append(enca.Children, child)
      offset += boxSize

      if h.Size == 0 {
         break
      }
   }
   return enca, nil
}

// Encode encodes the 'enca' box to a byte slice.
func (b *EncaBox) Encode() []byte {
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
