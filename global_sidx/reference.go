package sofia

import "154.pages.dev/sofia"

type Reference [3]uint32

func (Reference) Size() uint32 {
   return 3 * 4
}

// this is the size of the fragment, typically `moof` + `mdat`
func (r Reference) ReferencedSize() uint32 {
   return r[0] & sofia.Reference(r).Mask()
}

func (r Reference) SetReferencedSize(v uint32) {
   r[0] &= ^sofia.Reference(r).Mask()
   r[0] |= v
}
