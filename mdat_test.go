package sofia

import (
   "encoding/hex"
   "os"
   "testing"
)

func Test_Mdat(t *testing.T) {
   dst, err := os.Create("dec.m4v")
   if err != nil {
      t.Fatal(err)
   }
   defer dst.Close()
   init, err := os.ReadFile("testdata/amc-video/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   if _, err := dst.Write(init); err != nil {
      t.Fatal(err)
   }
   src, err := os.Open("testdata/amc-video/segment0.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var f File
   if err := f.Decode(src); err != nil {
      t.Fatal(err)
   }
   key, err := hex.DecodeString("c58d3308ed18d43776a78232f552dbe0")
   if err != nil {
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
   for _, b := range f.Moof.Traf.Boxes {
      if b.Header.Type() == "saiz" {
         b.Header.RawType = [4]byte{'f', 'r', 'e', 'e'}
      }
   }
   if err := f.Encode(dst); err != nil {
      t.Fatal(err)
   }
}
