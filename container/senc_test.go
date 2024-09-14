package container

import (
   "encoding/hex"
   "fmt"
   "os"
   "testing"
)

var senc_tests = []senc_test{
   {
      "../testdata/amc-avc1/init.m4f",
      "../testdata/amc-avc1/segment0.m4f",
      "c58d3308ed18d43776a78232f552dbe0",
      "amc-avc1.mp4",
   },
   {
      "../testdata/amc-mp4a/init.m4f",
      "../testdata/amc-mp4a/segment0.m4f",
      "91d888dfb0562ebc3abdd845d451e858",
      "amc-mp4a.mp4",
   },
   {
      "../testdata/hulu-avc1/init.mp4",
      "../testdata/hulu-avc1/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-avc1.mp4",
   },
   {
      "../testdata/hulu-ec-3/init.mp4",
      "../testdata/hulu-ec-3/segment-1.0001.m4s",
      "7be76f0d9c8a0db0b7f6059bf0a1c023",
      "hulu-ec-3.mp4",
   },
   {
      "../testdata/hulu-mp4a/init.mp4",
      "../testdata/hulu-mp4a/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-mp4a.mp4",
   },
   {
      "../testdata/mubi-avc1/video=300168.dash",
      "../testdata/mubi-avc1/video=300168-0.dash",
      "2556f746e8db3ee7f66fc22f5a28752a",
      "mubi-avc1.mp4",
   },
   {
      "../testdata/mubi-mp4a/audio_eng=268840.dash",
      "../testdata/mubi-mp4a/audio_eng=268840-0.dash",
      "2556f746e8db3ee7f66fc22f5a28752a",
      "mubi-mp4a.mp4",
   },
   {
      "../testdata/nbc-avc1/_227156876_5.mp4",
      "../testdata/nbc-avc1/_227156876_5_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-avc1.mp4",
   },
   {
      "../testdata/nbc-mp4a/_227156876_6_1.mp4",
      "../testdata/nbc-mp4a/_227156876_6_1_0.mp4",
      "3e2e8ccff89d0a72598a347feab5e7c8",
      "nbc-mp4a.mp4",
   },
   {
      "../testdata/paramount-avc1/0-17641.mp4",
      "../testdata/paramount-avc1/17642-196772.mp4",
      "efa0258cafde6102f513f031d0632290",
      "paramount-avc1.mp4",
   },
   {
      "../testdata/paramount-mp4a/init.m4v",
      "../testdata/paramount-mp4a/seg_1.m4s",
      "d98277ff6d7406ec398b49bbd52937d4",
      "paramount-mp4a.mp4",
   },
   {
      "../testdata/roku-avc1/index_video_8_0_init.mp4",
      "../testdata/roku-avc1/index_video_8_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-avc1.mp4",
   },
   {
      "../testdata/roku-mp4a/index_audio_2_0_init.mp4",
      "../testdata/roku-mp4a/index_audio_2_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-mp4a.mp4",
   },
}

type senc_test struct {
   init    string
   segment string
   key     string
   dst     string
}

func (s senc_test) encode_init() ([]byte, error) {
   buf, err := os.ReadFile(s.init)
   if err != nil {
      return nil, err
   }
   var value File
   err = value.Decode(buf)
   if err != nil {
      return nil, err
   }
   for _, pssh := range value.Moov.Pssh {
      copy(pssh.BoxHeader.Type[:], "free") // Firefox
   }
   description := value.Moov.Trak.Mdia.Minf.Stbl.Stsd
   if sinf, ok := description.Sinf(); ok {
      // Firefox
      copy(sinf.BoxHeader.Type[:], "free")
      if sample, ok := description.SampleEntry(); ok {
         // Firefox
         copy(sample.BoxHeader.Type[:], sinf.Frma.DataFormat[:])
      }
   }
   return value.Append(nil)
}

func (s senc_test) encode_segment(buf []byte) ([]byte, error) {
   fmt.Println(s.segment)
   segment, err := os.ReadFile(s.segment)
   if err != nil {
      return nil, err
   }
   var value File
   err = value.Decode(segment)
   if err != nil {
      return nil, err
   }
   track := value.Moof.Traf
   if senc := track.Senc; senc != nil {
      key, err := hex.DecodeString(s.key)
      if err != nil {
         return nil, err
      }
      for i, text := range value.Mdat.Data(&track) {
         err := senc.Sample[i].DecryptCenc(text, key)
         if err != nil {
            return nil, err
         }
      }
   }
   return value.Append(buf)
}

func TestSenc(t *testing.T) {
   for _, test := range senc_tests {
      buf, err := test.encode_init()
      if err != nil {
         t.Fatal(err)
      }
      buf, err = test.encode_segment(buf)
      if err != nil {
         t.Fatal(err)
      }
      err = os.WriteFile(test.dst, buf, os.ModePerm)
      if err != nil {
         t.Fatal(err)
      }
   }
}
