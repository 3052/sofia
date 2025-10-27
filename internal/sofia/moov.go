package mp4

import (
   "bytes"
   "encoding/binary"
   "errors"
)

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

func ParseMoov(data []byte) (MoovBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MoovBox{}, err
   }
   var moov MoovBox
   moov.Header = header
   moov.RawData = data[:header.Size]
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return MoovBox{}, errors.New("invalid child box size in moov")
      }
      childData := boxData[offset : offset+boxSize]
      var child MoovChild
      switch string(h.Type[:]) {
      case "trak":
         trak, err := ParseTrak(childData)
         if err != nil {
            return MoovBox{}, err
         }
         child.Trak = &trak
      case "pssh":
         pssh, err := ParsePssh(childData)
         if err != nil {
            return MoovBox{}, err
         }
         child.Pssh = &pssh
      default:
         child.Raw = childData
      }
      moov.Children = append(moov.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return moov, nil
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

// RemoveEncryption traverses the moov box and replaces encrypted sample entries.
func (b *MoovBox) RemoveEncryption() error {
   for _, trak := range b.GetAllTraks() {
      stsd := trak.GetStsd()
      if stsd == nil {
         continue
      }
      for i := range stsd.Children {
         stsdChild := &stsd.Children[i]
         if stsdChild.Encv != nil {
            newBoxData, err := b.rebuildVideoSampleEntry(stsdChild.Encv)
            if err != nil {
               return err
            }
            stsdChild.Encv = nil
            stsdChild.Raw = newBoxData
         } else if stsdChild.Enca != nil {
            newBoxData, err := b.rebuildAudioSampleEntry(stsdChild.Enca)
            if err != nil {
               return err
            }
            stsdChild.Enca = nil
            stsdChild.Raw = newBoxData
         }
      }
   }
   return nil
}

// RemoveDRM finds and renames all pssh boxes within this moov box to 'free'.
func (b *MoovBox) RemoveDRM() {
   for i := range b.Children {
      child := &b.Children[i]
      if child.Pssh != nil {
         child.Pssh.Header.Type = [4]byte{'f', 'r', 'e', 'e'}
      }
   }
}

func (b *MoovBox) rebuildVideoSampleEntry(encv *EncvBox) ([]byte, error) {
   var sinf *SinfBox
   for _, child := range encv.Children {
      if child.Sinf != nil {
         sinf = child.Sinf
         break
      }
   }
   if sinf == nil {
      return nil, errors.New("could not find 'sinf' box in encv")
   }
   var frma *FrmaBox
   for _, sinfChild := range sinf.Children {
      if f := sinfChild.Frma; f != nil {
         frma = f
         break
      }
   }
   if frma == nil {
      return nil, errors.New("could not find 'frma' box in sinf")
   }
   newFormatType := frma.DataFormat
   var newContent bytes.Buffer
   newContent.Write(encv.EntryHeader)
   for _, child := range encv.Children {
      if child.Sinf == nil {
         newContent.Write(child.Raw)
      }
   }
   newBoxSize := uint32(8 + newContent.Len())
   newBoxData := make([]byte, newBoxSize)
   binary.BigEndian.PutUint32(newBoxData[0:4], newBoxSize)
   copy(newBoxData[4:8], newFormatType[:])
   copy(newBoxData[8:], newContent.Bytes())
   return newBoxData, nil
}

func (b *MoovBox) rebuildAudioSampleEntry(enca *EncaBox) ([]byte, error) {
   var sinf *SinfBox
   for _, child := range enca.Children {
      if child.Sinf != nil {
         sinf = child.Sinf
         break
      }
   }
   if sinf == nil {
      return nil, errors.New("could not find 'sinf' box in enca")
   }
   var frma *FrmaBox
   for _, sinfChild := range sinf.Children {
      if f := sinfChild.Frma; f != nil {
         frma = f
         break
      }
   }
   if frma == nil {
      return nil, errors.New("could not find 'frma' box in sinf")
   }
   newFormatType := frma.DataFormat
   var newContent bytes.Buffer
   newContent.Write(enca.EntryHeader)
   for _, child := range enca.Children {
      if child.Sinf == nil {
         newContent.Write(child.Raw)
      }
   }
   newBoxSize := uint32(8 + newContent.Len())
   newBoxData := make([]byte, newBoxSize)
   binary.BigEndian.PutUint32(newBoxData[0:4], newBoxSize)
   copy(newBoxData[4:8], newFormatType[:])
   copy(newBoxData[8:], newContent.Bytes())
   return newBoxData, nil
}

func (b *MoovBox) GetTrakByTrackID(trackID uint32) *TrakBox {
   for _, child := range b.Children {
      if child.Trak != nil {
         if trackID == 1 {
            return child.Trak
         }
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
