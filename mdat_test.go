package sofia

import (
   "encoding/hex"
   "io"
   "os"
   "testing"
)

var tests = []testdata{
   {
      "testdata/nbc-video/_227156876_5.mp4",
      "testdata/nbc-video/_227156876_5_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-video.mp4",
   },
   {
      "testdata/amc-video/init.m4f",
      "testdata/amc-video/segment0.m4f",
      "c58d3308ed18d43776a78232f552dbe0",
      "amc-video.mp4",
   },
}

func Test_Mdat(t *testing.T) {
   for _, test := range tests {
      func() {
         dst, err := os.Create(test.out)
         if err != nil {
            t.Fatal(err)
         }
         defer dst.Close()
         if err := test.encode_init(dst); err != nil {
            t.Fatal(err)
         }
         if err := test.encode_segment(dst); err != nil {
            t.Fatal(err)
         }
      }()
      break
   }
}
type testdata struct {
   init string
   segment string
   key string
   out string
}

func (t testdata) encode_init(dst io.Writer) error {
   src, err := os.Open(t.init)
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
         //copy(b.Header.RawType[:], "free") // Firefox
      }
   }
   for _, entry := range f.Moov.Trak.Mdia.Minf.Stbl.Stsd.Entries {
      if entry.Entry.Header.Type() == "encv" {
         //copy(entry.Entry.Header.RawType[:], "avc1") // Firefox
         for _, b := range entry.Boxes {
            if b.Header.Type() == "sinf" {
               //copy(b.Header.RawType[:], "free") // Firefox
            }
         }
      }
   }
   return f.Encode(dst)
}

func (t testdata) encode_segment(dst io.Writer) error {
   src, err := os.Open(t.segment)
   if err != nil {
      return err
   }
   defer src.Close()
   var f File
   if err := f.Decode(src); err != nil {
      return err
   }
   key, err := hex.DecodeString(t.key)
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
