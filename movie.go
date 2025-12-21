package sofia

import (
   "bytes"
   "encoding/binary"
   "errors"
   "fmt"
)

// --- MOOV ---
type MoovChild struct {
   Mvhd *MvhdBox
   Trak *TrakBox
   Pssh *PsshBox
   Raw  []byte
}

type MoovBox struct {
   Header   BoxHeader
   Children []MoovChild
}

func (b *MoovBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child MoovChild
      switch string(header.Type[:]) {
      case "mvhd":
         var mvhd MvhdBox
         if err := mvhd.Parse(content); err != nil {
            return err
         }
         child.Mvhd = &mvhd
      case "trak":
         var trak TrakBox
         if err := trak.Parse(content); err != nil {
            return err
         }
         child.Trak = &trak
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

func (b *MoovBox) Encode() []byte {
   buffer := make([]byte, 8)
   for _, child := range b.Children {
      if child.Mvhd != nil {
         buffer = append(buffer, child.Mvhd.Encode()...)
      } else if child.Trak != nil {
         buffer = append(buffer, child.Trak.Encode()...)
      } else if child.Pssh != nil {
         // Skipped
      } else if child.Raw != nil {
         buffer = append(buffer, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *MoovBox) RemovePssh() {
   var kept []MoovChild
   for _, child := range b.Children {
      if child.Pssh != nil {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

func (b *MoovBox) RemoveMvex() {
   var kept []MoovChild
   for _, child := range b.Children {
      if len(child.Raw) >= 8 && string(child.Raw[4:8]) == "mvex" {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

func (b *MoovBox) Trak() (*TrakBox, bool) {
   for _, child := range b.Children {
      if child.Trak != nil {
         return child.Trak, true
      }
   }
   return nil, false
}

func (b *MoovBox) Mvhd() (*MvhdBox, bool) {
   for _, child := range b.Children {
      if child.Mvhd != nil {
         return child.Mvhd, true
      }
   }
   return nil, false
}

func (b *MoovBox) FindPssh(systemID []byte) (*PsshBox, bool) {
   for _, child := range b.Children {
      if child.Pssh != nil && bytes.Equal(child.Pssh.SystemID[:], systemID) {
         return child.Pssh, true
      }
   }
   return nil, false
}

// --- MVHD ---
type MvhdBox struct {
   Header           BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   RemainingData    []byte
}

func (b *MvhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return fmt.Errorf("mvhd box too small")
   }
   b.Version = data[8]
   copy(b.Flags[:], data[9:12])
   offset := 12
   if b.Version == 1 {
      if len(data) < 44 {
         return errors.New("mvhd v1 too short")
      }
      b.CreationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.ModificationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.Duration = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   } else { // Version 0
      if len(data) < 32 {
         return errors.New("mvhd v0 too short")
      }
      b.CreationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.ModificationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.Duration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   }
   if offset < int(b.Header.Size) {
      b.RemainingData = make([]byte, int(b.Header.Size)-offset)
      copy(b.RemainingData, data[offset:b.Header.Size])
   }
   return nil
}

func (b *MvhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

func (b *MvhdBox) Encode() []byte {
   var baseSize uint32
   if b.Version == 1 {
      baseSize = 44
   } else {
      baseSize = 32
   }
   totalSize := baseSize + uint32(len(b.RemainingData))
   buffer := make([]byte, totalSize)
   binary.BigEndian.PutUint32(buffer[0:4], totalSize)
   copy(buffer[4:8], b.Header.Type[:])
   buffer[8] = b.Version
   copy(buffer[9:12], b.Flags[:])
   offset := 12
   if b.Version == 1 {
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.CreationTime)
      offset += 8
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.ModificationTime)
      offset += 8
      binary.BigEndian.PutUint32(buffer[offset:offset+4], b.Timescale)
      offset += 4
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.Duration)
      offset += 8
   } else {
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.CreationTime))
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.ModificationTime))
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], b.Timescale)
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.Duration))
      offset += 4
   }
   copy(buffer[offset:], b.RemainingData)
   b.Header.Size = totalSize
   return buffer
}

// --- TRAK ---
type TrakChild struct {
   Mdia *MdiaBox
   Raw  []byte
}

type TrakBox struct {
   Header   BoxHeader
   Children []TrakChild
}

func (b *TrakBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child TrakChild
      switch string(header.Type[:]) {
      case "mdia":
         var mdia MdiaBox
         if err := mdia.Parse(content); err != nil {
            return err
         }
         child.Mdia = &mdia
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *TrakBox) Encode() []byte {
   buffer := make([]byte, 8)
   for _, child := range b.Children {
      if child.Mdia != nil {
         buffer = append(buffer, child.Mdia.Encode()...)
      } else if child.Raw != nil {
         buffer = append(buffer, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *TrakBox) RemoveEdts() {
   var kept []TrakChild
   for _, child := range b.Children {
      if len(child.Raw) >= 8 && string(child.Raw[4:8]) == "edts" {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

func (b *TrakBox) Mdia() (*MdiaBox, bool) {
   for _, child := range b.Children {
      if child.Mdia != nil {
         return child.Mdia, true
      }
   }
   return nil, false
}

// --- MDIA ---
type MdiaChild struct {
   Mdhd *MdhdBox
   Minf *MinfBox
   Raw  []byte
}

type MdiaBox struct {
   Header   BoxHeader
   Children []MdiaChild
}

func (b *MdiaBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child MdiaChild
      switch string(header.Type[:]) {
      case "mdhd":
         var mdhd MdhdBox
         if err := mdhd.Parse(content); err != nil {
            return err
         }
         child.Mdhd = &mdhd
      case "minf":
         var minf MinfBox
         if err := minf.Parse(content); err != nil {
            return err
         }
         child.Minf = &minf
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *MdiaBox) Encode() []byte {
   buffer := make([]byte, 8)
   for _, child := range b.Children {
      if child.Mdhd != nil {
         buffer = append(buffer, child.Mdhd.Encode()...)
      } else if child.Minf != nil {
         buffer = append(buffer, child.Minf.Encode()...)
      } else if child.Raw != nil {
         buffer = append(buffer, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *MdiaBox) MdhdRaw() ([]byte, bool) {
   for _, child := range b.Children {
      if child.Mdhd != nil {
         return child.Mdhd.Encode(), true
      }
   }
   return nil, false
}

func (b *MdiaBox) Mdhd() (*MdhdBox, bool) {
   for _, child := range b.Children {
      if child.Mdhd != nil {
         return child.Mdhd, true
      }
   }
   return nil, false
}

func (b *MdiaBox) Minf() (*MinfBox, bool) {
   for _, child := range b.Children {
      if child.Minf != nil {
         return child.Minf, true
      }
   }
   return nil, false
}

// --- MDHD ---
type MdhdBox struct {
   Header           BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   Language         [2]byte
   Quality          [2]byte
}

func (b *MdhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return fmt.Errorf("mdhd box too small")
   }
   b.Version = data[8]
   copy(b.Flags[:], data[9:12])
   offset := 12
   if b.Version == 1 {
      if len(data) < 44 {
         return errors.New("mdhd v1 too short")
      }
      b.CreationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.ModificationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.Duration = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   } else { // Version 0
      if len(data) < 32 {
         return errors.New("mdhd v0 too short")
      }
      b.CreationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.ModificationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.Duration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   }
   if len(data) < offset+4 {
      return errors.New("mdhd truncated at language/quality")
   }
   copy(b.Language[:], data[offset:offset+2])
   copy(b.Quality[:], data[offset+2:offset+4])
   return nil
}

func (b *MdhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

func (b *MdhdBox) Encode() []byte {
   var size uint32
   if b.Version == 1 {
      size = 44
   } else {
      size = 32
   }
   buffer := make([]byte, size)
   binary.BigEndian.PutUint32(buffer[0:4], size)
   copy(buffer[4:8], b.Header.Type[:])
   buffer[8] = b.Version
   copy(buffer[9:12], b.Flags[:])
   offset := 12
   if b.Version == 1 {
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.CreationTime)
      offset += 8
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.ModificationTime)
      offset += 8
      binary.BigEndian.PutUint32(buffer[offset:offset+4], b.Timescale)
      offset += 4
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.Duration)
      offset += 8
   } else {
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.CreationTime))
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.ModificationTime))
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], b.Timescale)
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.Duration))
      offset += 4
   }
   copy(buffer[offset:offset+2], b.Language[:])
   copy(buffer[offset+2:offset+4], b.Quality[:])
   b.Header.Size = size
   return buffer
}

// --- MINF ---
type MinfChild struct {
   Stbl *StblBox
   Raw  []byte
}

type MinfBox struct {
   Header   BoxHeader
   Children []MinfChild
}

func (b *MinfBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child MinfChild
      switch string(header.Type[:]) {
      case "stbl":
         var stbl StblBox
         if err := stbl.Parse(content); err != nil {
            return err
         }
         child.Stbl = &stbl
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *MinfBox) Encode() []byte {
   buffer := make([]byte, 8)
   for _, child := range b.Children {
      if child.Stbl != nil {
         buffer = append(buffer, child.Stbl.Encode()...)
      } else if child.Raw != nil {
         buffer = append(buffer, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *MinfBox) Stbl() (*StblBox, bool) {
   for _, child := range b.Children {
      if child.Stbl != nil {
         return child.Stbl, true
      }
   }
   return nil, false
}
