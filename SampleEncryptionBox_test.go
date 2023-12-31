package sofia

import (
   "encoding/hex"
   "io"
   "os"
   "testing"
)

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
   for _, b := range f.Movie.Boxes {
      if b.Header.BoxType() == "pssh" {
         copy(b.Header.Type[:], "free") // Firefox
      }
   }
   d := &f.Movie.Track.Media.MediaInformation.SampleTable.SampleDescription
   if d.AudioSample != nil {
      for _, b := range d.AudioSample.Boxes {
         if b.Header.BoxType() == "sinf" {
            copy(b.Header.Type[:], "free") // Firefox
         }
      }
      copy(
         d.AudioSample.Entry.Header.Type[:],
         d.AudioSample.ProtectionScheme.OriginalFormat.DataFormat[:],
      ) // Firefox
   }
   if d.VisualSample != nil {
      for _, b := range d.VisualSample.Boxes {
         if b.Header.BoxType() == "sinf" {
            copy(b.Header.Type[:], "free") // Firefox
         }
      }
      copy(d.VisualSample.Entry.Header.Type[:], "avc1") // Firefox
   }
   return f.Encode(dst)
}
func Test_SampleEncryption(t *testing.T) {
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

var tests = []testdata{
   {
      "testdata/amc-mp4a/init.m4f",
      "testdata/amc-mp4a/segment0.m4f",
      "91d888dfb0562ebc3abdd845d451e858",
      "amc-mp4a.mp4",
   },
   {
      "testdata/amc-avc1/init.m4f",
      "testdata/amc-avc1/segment0.m4f",
      "c58d3308ed18d43776a78232f552dbe0",
      "amc-avc1.mp4",
   },
   {
      "testdata/hulu-mp4a/init.mp4",
      "testdata/hulu-mp4a/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-mp4a.mp4",
   },
   {
      "testdata/hulu-avc1/init.mp4",
      "testdata/hulu-avc1/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-avc1.mp4",
   },
   {
      "testdata/nbc-mp4a/_227156876_6_1.mp4",
      "testdata/nbc-mp4a/_227156876_6_1_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-mp4a.mp4",
   },
   {
      "testdata/nbc-avc1/_227156876_5.mp4",
      "testdata/nbc-avc1/_227156876_5_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-avc1.mp4",
   },
   {
      "testdata/paramount-mp4a/init.m4v",
      "testdata/paramount-mp4a/seg_1.m4s",
      "d98277ff6d7406ec398b49bbd52937d4",
      "paramount-mp4a.mp4",
   },
   {
      "testdata/paramount-avc1/init.m4v",
      "testdata/paramount-avc1/seg_1.m4s",
      "d98277ff6d7406ec398b49bbd52937d4",
      "paramount-avc1.mp4",
   },
   {
      "testdata/roku-mp4a/index_audio_2_0_init.mp4",
      "testdata/roku-mp4a/index_audio_2_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-mp4a.mp4",
   },
   {
      "testdata/roku-avc1/index_video_8_0_init.mp4",
      "testdata/roku-avc1/index_video_8_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-avc1.mp4",
   },
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
   for i, data := range f.MediaData.Data {
      sample := f.MovieFragment.TrackFragment.SampleEncryption.Samples[i]
      err := sample.Decrypt_CENC(data, key)
      if err != nil {
         return err
      }
   }
   return f.Encode(dst)
}

type testdata struct {
   init string
   segment string
   key string
   out string
}
