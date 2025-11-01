package sofia

import (
   "encoding/binary"
   "errors"
)

type StsdChild struct {
   Encv *EncvBox
   Enca *EncaBox
   Raw  []byte
}

// Sinf finds the SinfBox within this sample entry.
// It returns the SinfBox, the header of the sample entry (e.g., 'encv'), and a boolean indicating if it's protected.
func (sc *StsdChild) Sinf() (sinf *SinfBox, sampleEntryHeader *BoxHeader, isProtected bool) {
   if sc.Encv != nil {
      for _, child := range sc.Encv.Children {
         if child.Sinf != nil {
            return child.Sinf, &sc.Encv.Header, true
         }
      }
   }
   if sc.Enca != nil {
      for _, child := range sc.Enca.Children {
         if child.Sinf != nil {
            return child.Sinf, &sc.Enca.Header, true
         }
      }
   }
   return nil, nil, false
}

type StsdBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []StsdChild
}

func (b *StsdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   entryCount := binary.BigEndian.Uint32(data[12:16])
   boxData := data[16:b.Header.Size]
   offset := 0
   for i := uint32(0); i < entryCount && offset < len(boxData); i++ {
      var h BoxHeader
      if err := h.Parse(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in stsd")
      }
      childData := boxData[offset : offset+boxSize]
      var child StsdChild
      switch string(h.Type[:]) {
      case "encv":
         var encv EncvBox
         if err := encv.Parse(childData); err != nil {
            return err
         }
         child.Encv = &encv
      case "enca":
         var enca EncaBox
         if err := enca.Parse(childData); err != nil {
            return err
         }
         child.Enca = &enca
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
   headerData := b.RawData[8:16]
   fullContent := append(headerData, content...)
   b.Header.Size = uint32(8 + len(fullContent))
   headerBytes := b.Header.Encode()
   return append(headerBytes, fullContent...)
}
