package sofia

import (
   "encoding/binary"
   "errors"
)

// FindMoov returns the first moov of a decoded box list.
func FindMoov(boxes []Box) (*MoovBox, bool) {
   for _, box := range boxes {
      if box.Moov != nil {
         return box.Moov, true
      }
   }
   return nil, false
}

// --- Box ---
type Box struct {
   Moov *MoovBox
   Moof *MoofBox
   Mdat *MdatBox
   Raw  []byte
}

func DecodeBoxes(data []byte) ([]Box, error) {
   var boxes []Box
   offset := 0
   for offset < len(data) {
      header, err := DecodeBoxHeader(data[offset:])
      if err != nil {
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
         moov, err := DecodeMoovBox(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Moov = moov
      case "moof":
         moof, err := DecodeMoofBox(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Moof = moof
      case "mdat":
         mdat, err := DecodeMdatBox(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Mdat = mdat
      default:
         currentBox.Raw = boxData
      }
      boxes = append(boxes, currentBox)
      offset += boxSize
   }
   return boxes, nil
}

// --- BoxHeader ---
type BoxHeader struct {
   Size uint32
   Type [4]byte
}

func DecodeBoxHeader(data []byte) (*BoxHeader, error) {
   if len(data) < 8 {
      return nil, errors.New("not enough data for box header")
   }
   h := &BoxHeader{}
   p := parser{data: data}
   h.Size = p.Uint32()
   copy(h.Type[:], p.Bytes(4))
   return h, nil
}

func (h *BoxHeader) Put(buffer []byte) {
   w := writer{buf: buffer}
   w.PutUint32(h.Size)
   w.PutBytes(h.Type[:])
}

// --- MDAT ---
type MdatBox struct {
   Header  *BoxHeader
   Payload []byte
}

func DecodeMdatBox(data []byte) (*MdatBox, error) {
   b := &MdatBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }
   size := int(b.Header.Size)
   if size == 0 { // box extends to the end of the data
      size = len(data)
   }
   if size < 8 || size > len(data) {
      return nil, errors.New("mdat box size out of range")
   }
   b.Payload = data[8:size]
   return b, nil
}

// --- READING HELPER ---

type parser struct {
   data   []byte
   offset int
}

func (p *parser) Byte() byte {
   val := p.data[p.offset]
   p.offset++
   return val
}

func (p *parser) Bytes(n int) []byte {
   val := p.data[p.offset : p.offset+n]
   p.offset += n
   return val
}

func (p *parser) Int32() int32 {
   val := int32(binary.BigEndian.Uint32(p.data[p.offset:]))
   p.offset += 4
   return val
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

func (p *parser) Uint64() uint64 {
   val := binary.BigEndian.Uint64(p.data[p.offset:])
   p.offset += 8
   return val
}

// --- WRITING HELPER ---

type writer struct {
   buf    []byte
   offset int
}

func (w *writer) PutByte(data byte) {
   w.buf[w.offset] = data
   w.offset++
}

func (w *writer) PutBytes(data []byte) {
   copy(w.buf[w.offset:], data)
   w.offset += len(data)
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

// box.go
