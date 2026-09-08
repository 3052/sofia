package sofia

import (
   "encoding/binary"
   "errors"
   "io"
)

// drain discards n bytes from r. n < 0 means discard to EOF.
func drain(r io.Reader, n int64) error {
   if n < 0 {
      _, err := io.Copy(io.Discard, r)
      return err
   }
   _, err := io.CopyN(io.Discard, r, n)
   return err
}

// readBoxHeader reads the next box header from r and returns the box type
// with the number of payload bytes that follow it. A payloadLen of -1
// means the box size field was 0 and the box extends to the end of the
// stream, which only makes sense for mdat.
func readBoxHeader(r io.Reader) (typ [4]byte, payloadLen int64, err error) {
   var h [8]byte
   if _, err = io.ReadFull(r, h[:]); err != nil {
      return
   }
   copy(typ[:], h[4:8])
   size := binary.BigEndian.Uint32(h[0:4])
   switch {
   case size == 0:
      payloadLen = -1
   case size == 1:
      var l [8]byte
      if _, err = io.ReadFull(r, l[:]); err != nil {
         return
      }
      n := int64(binary.BigEndian.Uint64(l[:]))
      if n < 16 {
         err = errors.New("invalid box largesize")
         return
      }
      payloadLen = n - 16
   case size < 8:
      err = errors.New("invalid box size")
   default:
      payloadLen = int64(size) - 8
   }
   return
}

// readWholeBox buffers a complete box (header included) so the []byte
// decoders can work on it unchanged. Only for small boxes: never mdat.
func readWholeBox(r io.Reader, typ [4]byte, payloadLen int64) ([]byte, error) {
   if payloadLen < 0 {
      return nil, errors.New("box extends to EOF; cannot buffer it")
   }
   buf := make([]byte, 8+payloadLen)
   binary.BigEndian.PutUint32(buf[0:4], uint32(8+payloadLen))
   copy(buf[4:8], typ[:])
   if _, err := io.ReadFull(r, buf[8:]); err != nil {
      return nil, err
   }
   return buf, nil
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

// countingReader tracks how many bytes have been consumed so Process can
// report a byte-exact resume point. The wrapped reader must not be
// buffered ahead of actual consumption (no bufio), or the count will run
// ahead of what has been processed.
type countingReader struct {
   R io.Reader
   N int64
}

func (c *countingReader) Read(p []byte) (int, error) {
   n, err := c.R.Read(p)
   c.N += int64(n)
   return n, err
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
