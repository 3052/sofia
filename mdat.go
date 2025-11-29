package sofia

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
