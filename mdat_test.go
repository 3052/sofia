package sofia

import (
   "encoding/hex"
   "os"
   "testing"
)

func Test_Mdat(t *testing.T) {
   src, err := os.Open("testdata/amc-video/segment0.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   key, err := hex.DecodeString("c58d3308ed18d43776a78232f552dbe0")
   if err != nil {
      t.Fatal(err)
   }
   var f File
   if err := f.Decode(src); err != nil {
      t.Fatal(err)
   }
   for i := range f.Mdat.Data {
      sample := f.Mdat.Data[i]
      enc := f.Moof.Traf.Senc.Samples[i]
      err := CryptSampleCenc(sample, key, enc)
      if err != nil {
         t.Fatal(err)
      }
   }
   init, err := os.ReadFile("testdata/amc-video/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   dst, err := os.Create("dec.m4v")
   if err != nil {
      t.Fatal(err)
   }
   defer dst.Close()
   dst.Write(init)
   if err := f.Encode(dst); err != nil {
      t.Fatal(err)
   }
}
