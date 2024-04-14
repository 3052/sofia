package sofia

import (
   "encoding/hex"
   "fmt"
   "io"
   "log/slog"
   "os"
   "testing"
)

func TestSampleEncryption(t *testing.T) {
   slog.SetLogLoggerLevel(slog.LevelDebug)
   tests = tests[:1]
   for _, test := range tests {
      func() {
         file, err := os.Create(test.out)
         if err != nil {
            t.Fatal(err)
         }
         defer file.Close()
         if err := test.encode_init(file); err != nil {
            t.Fatal(err)
         }
         return
         if err := test.encode_segment(file); err != nil {
            t.Fatal(err)
         }
      }()
   }
}

var tests = []testdata{
   {
      "testdata/tubi/0-30057.mp4",
      "",
      "",
      "tubi.mp4",
   },
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
   fmt.Println(t.init)
   src, err := os.Open(t.init)
   if err != nil {
      return err
   }
   defer src.Close()
   var value File
   if err := value.Read(src); err != nil {
      return err
   }
   for _, b := range value.Movie.Boxes {
      if b.BoxHeader.Type.String() == "pssh" { // moov
         copy(b.BoxHeader.Type[:], "free") // Firefox
      }
   }
   sample, protect := value.
      Movie.
      Track.
      Media.
      MediaInformation.
      SampleTable.
      SampleDescription.
      SampleEntry()
   // Firefox enca encv sinf
   copy(protect.BoxHeader.Type[:], "free")
   // Firefox stsd enca encv
   copy(sample.BoxHeader.Type[:], protect.OriginalFormat.DataFormat[:])
   return value.Write(dst)
}

func (t testdata) encode_segment(dst io.Writer) error {
   fmt.Println(t.segment)
   src, err := os.Open(t.segment)
   if err != nil {
      return err
   }
   defer src.Close()
   var file File
   if err := file.Read(src); err != nil {
      return err
   }
   key, err := hex.DecodeString(t.key)
   if err != nil {
      return err
   }
   fragment := file.MovieFragment.TrackFragment
   for i, data := range file.MediaData.Data(fragment.TrackRun) {
      err := fragment.SampleEncryption.Samples[i].DecryptCenc(data, key)
      if err != nil {
         return err
      }
   }
   return file.Write(dst)
}
