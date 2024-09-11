package container

import (
   "fmt"
   "os"
   "testing"
)

func TestTenc(t *testing.T) {
   src, err := os.Open("../testdata/amc-avc1/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var value File
   err = value.Read(src)
   if err != nil {
      t.Fatal(err)
   }
   sinf, ok := value.Moov.Trak.Mdia.Minf.Stbl.Stsd.Sinf()
   if !ok {
      t.Fatal("Sinf")
   }
   fmt.Printf("%+v\n", sinf.Schi.Tenc)
}
