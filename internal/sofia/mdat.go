package mp4

// MdatBox represents the 'mdat' box (Media Data Box).
type MdatBox struct {
   Header  BoxHeader
   Payload []byte
}

// Parse now correctly separates the header from the media payload.
func (b *MdatBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   // The payload is the content of the box *after* the 8-byte header.
   b.Payload = data[8:b.Header.Size]
   return nil
}

// Encode now correctly reconstructs the full box from the header and payload.
func (b *MdatBox) Encode() []byte {
   // The size in the header must be updated to reflect the current payload size.
   size := uint32(len(b.Payload) + 8)
   fullBox := make([]byte, size)
   b.Header.Size = size
   b.Header.Write(fullBox)
   copy(fullBox[8:], b.Payload)
   return fullBox
}
