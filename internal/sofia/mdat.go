package mp4

// MdatBox represents the 'mdat' box (Media Data Box).
type MdatBox struct {
   Header  BoxHeader
   Payload []byte
}

// ParseMdat now correctly separates the header from the media payload.
func ParseMdat(data []byte) (MdatBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MdatBox{}, err
   }
   var mdat MdatBox
   mdat.Header = header
   // The payload is the content of the box *after* the 8-byte header.
   mdat.Payload = data[8:header.Size]
   return mdat, nil
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
