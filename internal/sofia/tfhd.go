package mp4

import "encoding/binary"

// TfhdBox represents the 'tfhd' box (Track Fragment Header Box).
type TfhdBox struct {
   Header  BoxHeader
   Version byte
   Flags   uint32
   TrackID uint32
}

// ParseTfhd parses the 'tfhd' box from a byte slice.
func ParseTfhd(data []byte) (TfhdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TfhdBox{}, err
   }
   var tfhd TfhdBox
   tfhd.Header = header
   tfhd.Version = data[8]
   tfhd.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   tfhd.TrackID = binary.BigEndian.Uint32(data[12:16])
   return tfhd, nil
}

func (b *TfhdBox) Encode() []byte { return nil } // Omitted for brevity
