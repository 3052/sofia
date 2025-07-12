package file

import (
   "fmt"
   "os"
   "testing"
)

func TestSidx(t *testing.T) {
   data, err := os.ReadFile("../testdata/hulu-avc1/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   var fileVar File
   err = fileVar.Read(data)
   if err != nil {
      t.Fatal(err)
   }
   for _, reference := range fileVar.Sidx.Reference {
      fmt.Println(reference.Size())
   }
}
