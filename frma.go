package sofia

import "fmt"

type FrmaBox struct {
   Header     BoxHeader
   RawData    []byte
   DataFormat [4]byte
}

func (b *FrmaBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   // The dataFormat is the 4 bytes immediately following the box header.
   if len(data) < 12 {
      return fmt.Errorf("frma box is too small: %d bytes", len(data))
   }
   copy(b.DataFormat[:], data[8:12])
   return nil
}
