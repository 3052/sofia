package sofia

import (
   "encoding/hex"
   "io"
   "os"
   "testing"
)

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
   for i, data := range f.Media.Data {
      sample := f.MovieFragment.Track.Senc.Samples[i]
      err := sample.Decrypt_CENC(data, key)
      if err != nil {
         return err
      }
   }
   return f.Encode(dst)
}

var tests = []testdata{
   {
      "testdata/amc-audio/init.m4f",
      "testdata/amc-audio/segment0.m4f",
      "91d888dfb0562ebc3abdd845d451e858",
      "amc-audio.mp4",
   },
   {
      "testdata/amc-video/init.m4f",
      "testdata/amc-video/segment0.m4f",
      "c58d3308ed18d43776a78232f552dbe0",
      "amc-video.mp4",
   },
   {
      "testdata/hulu-audio/init.mp4",
      "testdata/hulu-audio/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-audio.mp4",
   },
   {
      "testdata/hulu-video/init.mp4",
      "testdata/hulu-video/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-video.mp4",
   },
   {
      "testdata/nbc-audio/_227156876_6_1.mp4",
      "testdata/nbc-audio/_227156876_6_1_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-audio.mp4",
   },
   {
      "testdata/nbc-video/_227156876_5.mp4",
      "testdata/nbc-video/_227156876_5_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-video.mp4",
   },
   {
      "testdata/paramount-audio/init.m4v",
      "testdata/paramount-audio/seg_1.m4s",
      "d98277ff6d7406ec398b49bbd52937d4",
      "paramount-audio.mp4",
   },
   {
      "testdata/paramount-video/init.m4v",
      "testdata/paramount-video/seg_1.m4s",
      "d98277ff6d7406ec398b49bbd52937d4",
      "paramount-video.mp4",
   },
   {
      "testdata/roku-audio/index_audio_2_0_init.mp4",
      "testdata/roku-audio/index_audio_2_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-audio.mp4",
   },
   {
      "testdata/roku-video/index_video_8_0_init.mp4",
      "testdata/roku-video/index_video_8_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-video.mp4",
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
   for _, b := range f.Movie.Boxes {
      if b.Header.BoxType() == "pssh" {
         copy(b.Header.Type[:], "free") // Firefox
      }
   }
   stsd := &f.Movie.Track.Mdia.Media.Sample.Stsd
   copy(stsd.Encv.Header.Type[:], "avc1") // Firefox
   for _, b := range stsd.Encv.Boxes {
      if b.Header.BoxType() == "sinf" {
         copy(b.Header.Type[:], "free") // Firefox
      }
   }
   for _, b := range stsd.Audio.Boxes {
      if b.Header.BoxType() == "sinf" {
         copy(b.Header.Type[:], "free") // Firefox
      }
   }
   copy(stsd.Audio.Header.Type[:], "mp4a") // Firefox
   return f.Encode(dst)
}

