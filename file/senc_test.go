package file

import (
   "encoding/hex"
   "fmt"
   "io"
   "os"
   "testing"
)

func (t testdata) encode_segment(write io.Writer) error {
   fmt.Println(t.segment)
   in, err := os.Open(t.segment)
   if err != nil {
      return err
   }
   defer in.Close()
   var out File
   err = out.Read(in)
   if err != nil {
      return err
   }
   track := out.MovieFragment.TrackFragment
   if encrypt := track.SampleEncryption; encrypt != nil {
      key, err := hex.DecodeString(t.key)
      if err != nil {
         return err
      }
      for i, text := range out.MediaData.Data(track) {
         err := encrypt.Sample[i].DecryptCenc(text, key)
         if err != nil {
            return err
         }
      }
   }
   return out.Write(write)
}

func TestSenc(t *testing.T) {
   for _, test := range tests {
      func() {
         out, err := os.Create(test.out)
         if err != nil {
            t.Fatal(err)
         }
         defer out.Close()
         err = test.encode_init(out)
         if err != nil {
            t.Fatal(err)
         }
         err = test.encode_segment(out)
         if err != nil {
            t.Fatal(err)
         }
      }()
   }
}

var tests = []testdata{
   {
      "testdata/amc-avc1/init.m4f",
      "testdata/amc-avc1/segment0.m4f",
      "c58d3308ed18d43776a78232f552dbe0",
      "amc-avc1.mp4",
   },
   {
      "testdata/amc-mp4a/init.m4f",
      "testdata/amc-mp4a/segment0.m4f",
      "91d888dfb0562ebc3abdd845d451e858",
      "amc-mp4a.mp4",
   },
   {
      "testdata/hulu-avc1/init.mp4",
      "testdata/hulu-avc1/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-avc1.mp4",
   },
   {
      "testdata/hulu-ec-3/init.mp4",
      "testdata/hulu-ec-3/segment-1.0001.m4s",
      "7be76f0d9c8a0db0b7f6059bf0a1c023",
      "hulu-ec-3.mp4",
   },
   {
      "testdata/hulu-mp4a/init.mp4",
      "testdata/hulu-mp4a/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-mp4a.mp4",
   },
   {
      "testdata/mubi-avc1/video=300168.dash",
      "testdata/mubi-avc1/video=300168-0.dash",
      "2556f746e8db3ee7f66fc22f5a28752a",
      "mubi-avc1.mp4",
   },
   {
      "testdata/mubi-mp4a/audio_eng=268840.dash",
      "testdata/mubi-mp4a/audio_eng=268840-0.dash",
      "2556f746e8db3ee7f66fc22f5a28752a",
      "mubi-mp4a.mp4",
   },
   {
      "testdata/nbc-avc1/_227156876_5.mp4",
      "testdata/nbc-avc1/_227156876_5_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-avc1.mp4",
   },
   {
      "testdata/nbc-mp4a/_227156876_6_1.mp4",
      "testdata/nbc-mp4a/_227156876_6_1_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-mp4a.mp4",
   },
   {
      "testdata/paramount-avc1/init.m4v",
      "testdata/paramount-avc1/seg_1.m4s",
      "efa0258cafde6102f513f031d0632290",
      "paramount-avc1.mp4",
   },
   {
      "testdata/paramount-mp4a/init.m4v",
      "testdata/paramount-mp4a/seg_1.m4s",
      "d98277ff6d7406ec398b49bbd52937d4",
      "paramount-mp4a.mp4",
   },
   {
      "testdata/roku-avc1/index_video_8_0_init.mp4",
      "testdata/roku-avc1/index_video_8_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-avc1.mp4",
   },
   {
      "testdata/roku-mp4a/index_audio_2_0_init.mp4",
      "testdata/roku-mp4a/index_audio_2_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-mp4a.mp4",
   },
}

type testdata struct {
   init    string
   segment string
   key     string
   out     string
}

func (t testdata) encode_init(dst io.Writer) error {
   src, err := os.Open(t.init)
   if err != nil {
      return err
   }
   defer src.Close()
   var value File
   err = value.Read(src)
   if err != nil {
      return err
   }
   for _, protect := range value.Movie.Protection {
      copy(protect.BoxHeader.Type[:], "free") // Firefox
   }
   description := value.Movie.
      Track.
      Media.
      MediaInformation.
      SampleTable.
      SampleDescription
   if protect, ok := description.Protection(); ok {
      // Firefox
      copy(protect.BoxHeader.Type[:], "free")
      if sample, ok := description.SampleEntry(); ok {
         // Firefox
         copy(sample.BoxHeader.Type[:], protect.OriginalFormat.DataFormat[:])
      }
   }
   return value.Write(dst)
}
