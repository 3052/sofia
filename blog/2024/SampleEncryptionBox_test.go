package sofia

import (
   "154.pages.dev/sofia"
   "encoding/hex"
   "io"
   "os"
   "testing"
)

type testdata struct {
   init string
   segment string
   key string
   out string
}

func Test_SampleEncryption(t *testing.T) {
   test := testdata{
      "init.m4v",
      "seg_1.m4s",
      "efa0258cafde6102f513f031d0632290",
      "out.m4v",
   }
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
   if err := test.Segment_2(dst); err != nil {
      t.Fatal(err)
   }
}

func (t testdata) Segment_2(dst io.Writer) error {
   src, err := os.Open("seg_2.m4s")
   if err != nil {
      return err
   }
   defer src.Close()
   var f sofia.File
   if err := f.Decode(src); err != nil {
      return err
   }
   key, err := hex.DecodeString(t.key)
   if err != nil {
      return err
   }
   for i, data := range f.MediaData.Data {
      sample := f.MovieFragment.TrackFragment.SampleEncryption.Samples[i]
      err := sample.Decrypt_CENC(data, key)
      if err != nil {
         return err
      }
      if _, err := dst.Write(data); err != nil {
         return err
      }
   }
   return nil
}

func (t testdata) encode_segment(dst io.Writer) error {
   src, err := os.Open(t.segment)
   if err != nil {
      return err
   }
   defer src.Close()
   var f sofia.File
   if err := f.Decode(src); err != nil {
      return err
   }
   key, err := hex.DecodeString(t.key)
   if err != nil {
      return err
   }
   for i, data := range f.MediaData.Data {
      sample := f.MovieFragment.TrackFragment.SampleEncryption.Samples[i]
      err := sample.Decrypt_CENC(data, key)
      if err != nil {
         return err
      }
   }
   f.MediaData.Header.Size = 0
   return f.Encode(dst)
}

func (t testdata) encode_init(dst io.Writer) error {
   data, err := os.ReadFile(t.init)
   if err != nil {
      return err
   }
   if _, err := dst.Write(data); err != nil {
      return err
   }
   return nil
}
