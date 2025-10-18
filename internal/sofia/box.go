package mp4parser

import "fmt"

// Box is a container for a parsed top-level MP4 box.
type Box struct {
   Header *BoxHeader
   Moof   *MoofBox
   Mdat   *MdatBox
}

// Size calculates the total size of the contained top-level box.
func (b *Box) Size() uint64 {
   if b.Moof != nil {
      return b.Moof.Size()
   }
   if b.Mdat != nil {
      return b.Mdat.Size()
   }
   return 0
}

// Format serializes the top-level box into a new byte slice.
func (b *Box) Format() ([]byte, error) {
   size := b.Size()
   if size == 0 {
      return nil, fmt.Errorf("box is empty, cannot format")
   }
   dst := make([]byte, size)
   if b.Moof != nil {
      b.Moof.Format(dst, 0)
   }
   if b.Mdat != nil {
      b.Mdat.Format(dst, 0)
   }
   return dst, nil
}
