package sofia

import "errors"

type FrmaBox struct {
   Header     BoxHeader
   DataFormat [4]byte
}

func (b *FrmaBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return errors.New("frma box is too small")
   }
   copy(b.DataFormat[:], data[8:12])
   return nil
}
