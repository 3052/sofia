package sofia

type EdtsBox struct {
   Header  BoxHeader
   RawData []byte
}

func (b *EdtsBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   return nil
}

func (b *EdtsBox) Encode() []byte {
   content := b.RawData[8:]
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}
