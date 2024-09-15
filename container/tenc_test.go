package container

import (
   "fmt"
   "os"
   "testing"
)

func TestTenc(t *testing.T) {
   buf, err := os.ReadFile("../testdata/amc-avc1/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   var value File
   err = value.Read(buf)
   if err != nil {
      t.Fatal(err)
   }
   sinf, ok := value.Moov.Trak.Mdia.Minf.Stbl.Stsd.Sinf()
   if !ok {
      t.Fatal("Sinf")
   }
   fmt.Printf("%+v\n", sinf.Schi.Tenc)
}
