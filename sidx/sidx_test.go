package sidx

import (
   "154.pages.dev/sofia"
   "fmt"
   "testing"
)

func TestDecode(t *testing.T) {
   var ref Reference
   ref.SetSize(9)
   var index Box
   copy(index.BoxHeader.Type[:], "sidx")
   index.ReferenceCount = 1
   index.EarliestPresentationTime = make([]byte, 4)
   index.FirstOffset = make([]byte, 4)
   index.Reference = []Reference{ref}
   index.BoxHeader.Size = uint32(index.GetSize())
   buf, err := index.Append(nil)
   if err != nil {
      t.Fatal(err)
   }
   var head sofia.BoxHeader
   n, err := head.Decode(buf)
   if err != nil {
      t.Fatal(err)
   }
   index = Box{BoxHeader: head}
   err = index.Read(buf[n:])
   if err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", index)
}

func TestReference(t *testing.T) {
   var ref Reference
   ref.SetSize(9)
   fmt.Println(ref)
}
