package container

import (
   "fmt"
   "os"
   "testing"
)

func TestTenc(t *testing.T) {
   data, err := os.ReadFile("../testdata/amc-avc1/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   var file1 File
   err = file1.Read(data)
   if err != nil {
      t.Fatal(err)
   }
   sinf, ok := file1.Moov.Trak.Mdia.Minf.Stbl.Stsd.Sinf()
   if !ok {
      t.Fatal("Sinf")
   }
   fmt.Printf("%+v\n", sinf.Schi.Tenc)
}
