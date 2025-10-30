package sofia

import "errors"

type MoovChild struct {
   Trak *TrakBox
   Pssh *PsshBox
   Raw  []byte
}

type MoovBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []MoovChild
}

func (b *MoovBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if err := h.Parse(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in moov")
      }
      childData := boxData[offset : offset+boxSize]
      var child MoovChild
      switch string(h.Type[:]) {
      case "trak":
         var trak TrakBox
         if err := trak.Parse(childData); err != nil {
            return err
         }
         child.Trak = &trak
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(childData); err != nil {
            return err
         }
         child.Pssh = &pssh
      default:
         child.Raw = childData
      }
      b.Children = append(b.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return nil
}

func (b *MoovBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Trak != nil {
         content = append(content, child.Trak.Encode()...)
      } else if child.Pssh != nil {
         content = append(content, child.Pssh.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

// Sanitize removes all DRM and encryption signaling from the moov box.
// It renames pssh boxes to 'free' and updates encrypted sample entries.
func (b *MoovBox) Sanitize() error {
   // Rename top-level pssh boxes within this moov box to 'free'.
   for i := range b.Children {
      child := &b.Children[i]
      if child.Pssh != nil {
         child.Pssh.Header.Type = [4]byte{'f', 'r', 'e', 'e'}
      }
   }

   // Get the single track and sanitize it.
   if trak, ok := b.GetTrak(); ok {
      stsd := trak.GetStsd()
      if stsd == nil {
         return nil // No sample descriptions to sanitize.
      }
      for i := range stsd.Children {
         stsdChild := &stsd.Children[i]

         sampleEntryHeader, sinf, isEncrypted := stsdChild.GetEncryptionInfo()
         if !isEncrypted {
            continue
         }

         if sinf == nil {
            return errors.New("could not find 'sinf' box to remove")
         }

         frma := sinf.GetFrma()
         if frma == nil {
            return errors.New("could not find 'frma' box for original format")
         }

         // Perform the sanitization.
         sinf.Header.Type = [4]byte{'f', 'r', 'e', 'e'}
         sampleEntryHeader.Type = frma.DataFormat
      }
   }
   return nil
}

// GetTrak returns the first trak box found and a boolean indicating if it was found.
func (b *MoovBox) GetTrak() (*TrakBox, bool) {
   for _, child := range b.Children {
      if child.Trak != nil {
         return child.Trak, true
      }
   }
   return nil, false
}

func (b *MoovBox) AllPssh() []*PsshBox {
   var psshBoxes []*PsshBox
   for _, child := range b.Children {
      if child.Pssh != nil {
         psshBoxes = append(psshBoxes, child.Pssh)
      }
   }
   return psshBoxes
}
