package sofia

import (
   "encoding/binary"
   "errors"
)

// --- MOOF ---
type MoofChild struct {
   Traf *TrafBox
   Pssh *PsshBox
   Raw  []byte
}

type MoofBox struct {
   Header   BoxHeader
   Children []MoofChild
}

func (b *MoofBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child MoofChild
      switch string(h.Type[:]) {
      case "traf":
         var traf TrafBox
         if err := traf.Parse(content); err != nil {
            return err
         }
         child.Traf = &traf
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(content); err != nil {
            return err
         }
         child.Pssh = &pssh
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *MoofBox) Traf() (*TrafBox, bool) {
   for _, child := range b.Children {
      if child.Traf != nil {
         return child.Traf, true
      }
   }
   return nil, false
}

// --- TRAF ---
type TrafChild struct {
   Tfhd *TfhdBox
   Trun *TrunBox
   Senc *SencBox
   Raw  []byte
}

type TrafBox struct {
   Header   BoxHeader
   Children []TrafChild
}

func (b *TrafBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child TrafChild
      switch string(h.Type[:]) {
      case "tfhd":
         var tfhd TfhdBox
         if err := tfhd.Parse(content); err != nil {
            return err
         }
         child.Tfhd = &tfhd
      case "trun":
         var trun TrunBox
         if err := trun.Parse(content); err != nil {
            return err
         }
         child.Trun = &trun
      case "senc":
         var senc SencBox
         if err := senc.Parse(content); err != nil {
            return err
         }
         child.Senc = &senc
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *TrafBox) Tfhd() *TfhdBox {
   for _, child := range b.Children {
      if child.Tfhd != nil {
         return child.Tfhd
      }
   }
   return nil
}

func (b *TrafBox) Trun() *TrunBox {
   for _, child := range b.Children {
      if child.Trun != nil {
         return child.Trun
      }
   }
   return nil
}

func (b *TrafBox) Senc() (*SencBox, bool) {
   for _, child := range b.Children {
      if child.Senc != nil {
         return child.Senc, true
      }
   }
   return nil, false
}

// --- TFHD ---
type TfhdBox struct {
   Header                 BoxHeader
   Flags                  uint32
   TrackID                uint32
   BaseDataOffset         uint64
   SampleDescriptionIndex uint32
   DefaultSampleDuration  uint32
   DefaultSampleSize      uint32
   DefaultSampleFlags     uint32
}

func (b *TfhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   b.TrackID = binary.BigEndian.Uint32(data[12:16])
   offset := 16
   if b.Flags&0x000001 != 0 {
      if offset+8 > len(data) {
         return errors.New("tfhd too short")
      }
      b.BaseDataOffset = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   }
   if b.Flags&0x000002 != 0 {
      if offset+4 > len(data) {
         return errors.New("tfhd too short")
      }
      b.SampleDescriptionIndex = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if b.Flags&0x000008 != 0 {
      if offset+4 > len(data) {
         return errors.New("tfhd too short")
      }
      b.DefaultSampleDuration = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if b.Flags&0x000010 != 0 {
      if offset+4 > len(data) {
         return errors.New("tfhd too short")
      }
      b.DefaultSampleSize = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if b.Flags&0x000020 != 0 {
      if offset+4 > len(data) {
         return errors.New("tfhd too short")
      }
      b.DefaultSampleFlags = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   return nil
}

// --- TRUN ---
type SampleInfo struct {
   Size                  uint32
   Duration              uint32
   Flags                 uint32
   CompositionTimeOffset int32
}

type TrunBox struct {
   Header           BoxHeader
   Flags            uint32
   SampleCount      uint32
   DataOffset       int32
   FirstSampleFlags uint32
   Samples          []SampleInfo
}

func (b *TrunBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   b.SampleCount = binary.BigEndian.Uint32(data[12:16])
   offset := 16
   // Data Offset Present
   if b.Flags&0x000001 != 0 {
      if offset+4 > len(data) {
         return errors.New("trun too short for data offset")
      }
      b.DataOffset = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   }
   // First Sample Flags Present (0x04)
   if b.Flags&0x000004 != 0 {
      if offset+4 > len(data) {
         return errors.New("trun too short for first sample flags")
      }
      b.FirstSampleFlags = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   // Calculate size of one sample entry
   sampleEntrySize := 0
   if b.Flags&0x000100 != 0 {
      sampleEntrySize += 4
   } // Duration
   if b.Flags&0x000200 != 0 {
      sampleEntrySize += 4
   } // Size
   if b.Flags&0x000400 != 0 {
      sampleEntrySize += 4
   } // Flags
   if b.Flags&0x000800 != 0 {
      sampleEntrySize += 4
   } // CTO
   // Safety check
   if len(data)-offset < int(b.SampleCount)*sampleEntrySize {
      return errors.New("trun box too short for declared samples")
   }
   b.Samples = make([]SampleInfo, b.SampleCount)
   for i := uint32(0); i < b.SampleCount; i++ {
      if b.Flags&0x000100 != 0 {
         b.Samples[i].Duration = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if b.Flags&0x000200 != 0 {
         b.Samples[i].Size = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if b.Flags&0x000400 != 0 {
         b.Samples[i].Flags = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if b.Flags&0x000800 != 0 {
         val := binary.BigEndian.Uint32(data[offset : offset+4])
         b.Samples[i].CompositionTimeOffset = int32(val)
         offset += 4
      }
   }
   return nil
}
