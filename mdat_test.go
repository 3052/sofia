package sofia

import (
   "encoding/hex"
   "io"
   "os"
   "testing"
)

func encode_segment(dst io.Writer) error {
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

func encode_init(dst io.Writer) error {
   src, err := os.Open("testdata/amc-video/init.m4f")
   if err != nil {
      return err
   }
   defer src.Close()
   var f File
   if err := f.Decode(src); err != nil {
      return err
   }
   for _, b := range f.Moov.Boxes {
      if b.Header.Type() == "pssh" {
         copy(b.Header.RawType[:], "free") // Firefox
      }
   }
   for _, entry := range f.Moov.Trak.Mdia.Minf.Stbl.Stsd.Entries {
      if entry.Entry.Header.Type() == "encv" {
         copy(entry.Entry.Header.RawType[:], "avc1") // Firefox
         for _, b := range entry.Boxes {
            if b.Header.Type() == "sinf" {
               copy(b.Header.RawType[:], "free") // Firefox
            }
         }
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
   if err := encode_init(dst); err != nil {
      t.Fatal(err)
   }
   if err := encode_segment(dst); err != nil {
      t.Fatal(err)
   }
}
