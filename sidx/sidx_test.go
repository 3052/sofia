package sidx

import (
   "os"
   "testing"
)

func TestDecode(t *testing.T) {
   var ref Reference
   ref.SetSize(9)
   var index Box
   copy(index.BoxHeader.Type[:], "sidx")
   index.EarliestPresentationTime = make([]byte, 4)
   index.FirstOffset = make([]byte, 4)
   index.Reference = []Reference{ref}
   index.ReferenceCount = 1
   index.BoxHeader.Size = uint32(index.GetSize())
   buf, err := index.Append(nil)
   if err != nil {
      t.Fatal(err)
   }
   os.WriteFile("reverse.txt", buf, os.ModePerm)
}
