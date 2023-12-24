package sofia

import (
   "encoding/hex"
   "io"
   "os"
   "testing"
)

func segment(dst io.Writer) error {
   src, err := os.Open("testdata/amc-video/segment0.m4f")
   if err != nil {
      return err
   }
   defer src.Close()
   var f File
   if err := f.Decode(src); err != nil {
      return err
   }
   key, err := hex.DecodeString("c58d3308ed18d43776a78232f552dbe0")
   if err != nil {
      return err
   }
   for i := range f.Mdat.Data {
      sample := f.Mdat.Data[i]
      enc := f.Moof.Traf.Senc.Samples[i]
      err := CryptSampleCenc(sample, key, enc)
      if err != nil {
         return err
      }
   }
   return f.Encode(dst)
}

func Test_Mdat(t *testing.T) {
   dst, err := os.Create("dec.m4v")
   if err != nil {
      t.Fatal(err)
   }
   defer dst.Close()
   src, err := os.Open("testdata/amc-video/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var f File
   if err := f.Decode(src); err != nil {
      t.Fatal(err)
   }
   for _, b := range f.Moov.Boxes {
      if b.Header.Type() == "pssh" {
         // Firefox
         copy(b.Header.RawType[:], "free")
      }
   }
   for _, entry := range f.Moov.Trak.Mdia.Minf.Stbl.Stsd.Entries {
      if entry.Header.Type() == "encv" {
         // Firefox
         copy(entry.Header.RawType[:], "avc1")
         for _, b := range entry.Boxes {
            if b.Header.Type() == "sinf" {
               // Firefox
               copy(b.Header.RawType[:], "free")
            }
         }
      }
   }
   if err := f.Encode(dst); err != nil {
      t.Fatal(err)
   }
   if err := segment(dst); err != nil {
      t.Fatal(err)
   }
}

/*
firefox
encv -> avc1, sinf -> free
[moov] Size=1948
  [trak] Size=576
    [mdia] Size=476
      [minf] Size=383
        [stbl] Size=319
          [stsd] Size=243 Version=0 Flags=0x000000 EntryCount=1
            [encv] Size=227 ... (use "-full encv" to show all)
              [sinf] Size=80
*/
