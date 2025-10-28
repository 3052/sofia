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

// Parse parses the 'moov' box from a byte slice.
func (b *MoovBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if _, err := h.Read(boxData[offset:]); err != nil {
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
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
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

   // Traverse into each track to remove encryption signaling.
   for _, trak := range b.GetAllTraks() {
      stsd := trak.GetStsd()
      if stsd == nil {
         continue
      }
      for i := range stsd.Children {
         stsdChild := &stsd.Children[i]

         var sampleEntryHeader *BoxHeader
         var sinf *SinfBox
         var isEncrypted bool

         if stsdChild.Encv != nil {
            isEncrypted = true
            sampleEntryHeader = &stsdChild.Encv.Header
            for j := range stsdChild.Encv.Children {
               if stsdChild.Encv.Children[j].Sinf != nil {
                  sinf = stsdChild.Encv.Children[j].Sinf
                  break
               }
            }
         } else if stsdChild.Enca != nil {
            isEncrypted = true
            sampleEntryHeader = &stsdChild.Enca.Header
            for j := range stsdChild.Enca.Children {
               if stsdChild.Enca.Children[j].Sinf != nil {
                  sinf = stsdChild.Enca.Children[j].Sinf
                  break
               }
            }
         }

         if !isEncrypted {
            continue
         }

         if sinf == nil {
            return errors.New("could not find 'sinf' box to remove")
         }
         var frma *FrmaBox
         for _, sinfChild := range sinf.Children {
            if f := sinfChild.Frma; f != nil {
               frma = f
               break
            }
         }
         if frma == nil {
            return errors.New("could not find 'frma' box for original format")
         }

         sinf.Header.Type = [4]byte{'f', 'r', 'e', 'e'}
         sampleEntryHeader.Type = frma.DataFormat
      }
   }
   return nil
}

// GetTrak returns the first trak box found, assuming there is only one.
func (b *MoovBox) GetTrak() *TrakBox {
   for _, child := range b.Children {
      if child.Trak != nil {
         return child.Trak
      }
   }
   return nil
}

func (b *MoovBox) GetAllTraks() []*TrakBox {
   var traks []*TrakBox
   for _, child := range b.Children {
      if child.Trak != nil {
         traks = append(traks, child.Trak)
      }
   }
   return traks
}
