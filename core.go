package sofia

import (
   "encoding/binary"
   "errors"
)

// --- READING HELPER ---

type parser struct {
   data   []byte
   offset int
}

func (p *parser) Uint16() uint16 {
   val := binary.BigEndian.Uint16(p.data[p.offset:])
   p.offset += 2
   return val
}

func (p *parser) Uint32() uint32 {
   val := binary.BigEndian.Uint32(p.data[p.offset:])
   p.offset += 4
   return val
}

func (p *parser) Int32() int32 {
   val := int32(binary.BigEndian.Uint32(p.data[p.offset:]))
   p.offset += 4
   return val
}

func (p *parser) Uint64() uint64 {
   val := binary.BigEndian.Uint64(p.data[p.offset:])
   p.offset += 8
   return val
}

func (p *parser) Bytes(n int) []byte {
   val := p.data[p.offset : p.offset+n]
   p.offset += n
   return val
}

// --- WRITING HELPER ---

type writer struct {
   buf    []byte
   offset int
}

func (w *writer) PutUint16(val uint16) {
   binary.BigEndian.PutUint16(w.buf[w.offset:], val)
   w.offset += 2
}

func (w *writer) PutUint32(val uint32) {
   binary.BigEndian.PutUint32(w.buf[w.offset:], val)
   w.offset += 4
}

func (w *writer) PutUint64(val uint64) {
   binary.BigEndian.PutUint64(w.buf[w.offset:], val)
   w.offset += 8
}

func (w *writer) PutBytes(data []byte) {
   copy(w.buf[w.offset:], data)
   w.offset += len(data)
}

func (w *writer) PutByte(data byte) {
   w.buf[w.offset] = data
   w.offset++
}

// --- BoxHeader ---
type BoxHeader struct {
   Size uint32
   Type [4]byte
}

func (h *BoxHeader) Parse(data []byte) error {
   if len(data) < 8 {
      return errors.New("not enough data for box header")
   }
   p := parser{data: data}
   h.Size = p.Uint32()
   copy(h.Type[:], p.Bytes(4))
   return nil
}

func (h *BoxHeader) Put(buffer []byte) {
   w := writer{buf: buffer}
   w.PutUint32(h.Size)
   w.PutBytes(h.Type[:])
}

// --- Box ---
type Box struct {
   Moov *MoovBox
   Moof *MoofBox
   Mdat *MdatBox
   Sidx *SidxBox
   Pssh *PsshBox
   Raw  []byte
}

func (b *Box) Encode() []byte {
   switch {
   case b.Moov != nil:
      return b.Moov.Encode()
   default:
      return b.Raw
   }
}

func Parse(data []byte) ([]Box, error) {
   var boxes []Box
   offset := 0
   for offset < len(data) {
      var header BoxHeader
      if err := header.Parse(data[offset:]); err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(data) - offset
      }
      if boxSize < 8 || offset+boxSize > len(data) {
         return nil, errors.New("invalid child box size")
      }

      boxData := data[offset : offset+boxSize]
      var currentBox Box
      switch string(header.Type[:]) {
      case "moov":
         var moov MoovBox
         if err := moov.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Moov = &moov
      case "moof":
         var moof MoofBox
         if err := moof.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Moof = &moof
      case "mdat":
         var mdat MdatBox
         if err := mdat.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Mdat = &mdat
      case "sidx":
         var sidx SidxBox
         if err := sidx.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Sidx = &sidx
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(boxData); err != nil {
            return nil, err
         }
         currentBox.Pssh = &pssh
      default:
         currentBox.Raw = boxData
      }
      boxes = append(boxes, currentBox)
      offset += boxSize
   }
   return boxes, nil
}

// --- Finders ---
func FindMoov(boxes []Box) (*MoovBox, bool) {
   for _, box := range boxes {
      if box.Moov != nil {
         return box.Moov, true
      }
   }
   return nil, false
}

func FindSidx(boxes []Box) (*SidxBox, bool) {
   for _, box := range boxes {
      if box.Sidx != nil {
         return box.Sidx, true
      }
   }
   return nil, false
}

// --- MDAT ---
type MdatBox struct {
   Header  BoxHeader
   Payload []byte
}

func (b *MdatBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.Payload = data[8:b.Header.Size]
   return nil
}

// --- SIDX ---
type SidxReference struct {
   ReferenceType      bool
   ReferencedSize     uint32
   SubsegmentDuration uint32
   StartsWithSAP      bool
   SAPType            uint8
   SAPDeltaTime       uint32
}

type SidxBox struct {
   Header                   BoxHeader
   Version                  byte
   Flags                    uint32
   ReferenceID              uint32
   Timescale                uint32
   EarliestPresentationTime uint64
   FirstOffset              uint64
   References               []SidxReference
}

func (b *SidxBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 20 { // 8 byte header + 12 bytes of fields before version check
      return errors.New("sidx box too short")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Uint32()
   b.Version = byte(versionAndFlags >> 24)
   b.Flags = versionAndFlags & 0x00FFFFFF
   b.ReferenceID = p.Uint32()
   b.Timescale = p.Uint32()

   if b.Version == 0 {
      if len(data) < p.offset+8 {
         return errors.New("sidx v0 box too short")
      }
      b.EarliestPresentationTime = uint64(p.Uint32())
      b.FirstOffset = uint64(p.Uint32())
   } else {
      if len(data) < p.offset+16 {
         return errors.New("sidx v1 box too short")
      }
      b.EarliestPresentationTime = p.Uint64()
      b.FirstOffset = p.Uint64()
   }

   if len(data) < p.offset+4 {
      return errors.New("sidx box too short for reference_count")
   }
   _ = p.Uint16() // reserved
   referenceCount := p.Uint16()

   if len(data)-p.offset < int(referenceCount)*12 {
      return errors.New("sidx box too short for declared references")
   }

   b.References = make([]SidxReference, referenceCount)
   for i := 0; i < int(referenceCount); i++ {
      val1 := p.Uint32()
      b.References[i].ReferenceType = (val1 >> 31) == 1
      b.References[i].ReferencedSize = val1 & 0x7FFFFFFF

      b.References[i].SubsegmentDuration = p.Uint32()

      val2 := p.Uint32()
      b.References[i].StartsWithSAP = (val2 >> 31) == 1
      b.References[i].SAPType = uint8((val2 >> 28) & 0x07)
      b.References[i].SAPDeltaTime = val2 & 0x0FFFFFFF
   }
   return nil
}
