package file

import (
   "encoding/hex"
   "fmt"
   "os"
   "testing"
)

const folder = "../testdata/"

var senc_tests = []senc_test{
   {
      "criterion-avc1/0-804.mp4",
      "criterion-avc1/13845-168166.mp4",
      "377772323b0f45efb2c53c603749d834",
      "criterion-avc1.mp4",
   },
   {
      "hboMax-dvh1/0-902.mp4"
      "hboMax-dvh1/903-28882.mp4",
      "ee0d569c019057569eaf28b988c206f6",
      "hboMax-dvh1.mp4",
   },
   // hboMax-ec-3\bytes=0-19985.mp4
   // hboMax-ec-3\bytes=19986-149146.mp4
   // 
   // hboMax-hvc1\init.mp4
   // hboMax-hvc1\segment-1.0001.m4s
   //
   // hulu-avc1\map.mp4
   // hulu-avc1\pts_0.mp4
   {
      "paramount-mp4a/init.m4v",
      "paramount-mp4a/seg_1.m4s",
      "d98277ff6d7406ec398b49bbd52937d4",
      "paramount-mp4a.mp4",
   },
   {
      "roku-avc1/index_video_8_0_init.mp4",
      "roku-avc1/index_video_8_0_1.mp4",
      "1ba08384626f9523e37b9db17f44da2b",
      "roku-avc1.mp4",
   },
   // rtbf-avc1\vod-idx-2-video=300000-0.dash
   // rtbf-avc1\vod-idx-2-video=300000.dash
   //
   // tubi-avc1\0-30057.mp4
   // tubi-avc1\30058-111481.mp4
}

func (s *senc_test) encode_init() ([]byte, error) {
   data, err := os.ReadFile(s.init)
   if err != nil {
      return nil, err
   }
   var file1 File
   err = file1.Read(data)
   if err != nil {
      return nil, err
   }
   for _, pssh := range file1.Moov.Pssh {
      copy(pssh.BoxHeader.Type[:], "free") // Firefox
   }
   description := file1.Moov.Trak.Mdia.Minf.Stbl.Stsd
   if sinf, ok := description.Sinf(); ok {
      // Firefox
      copy(sinf.BoxHeader.Type[:], "free")
      if sample, ok := description.SampleEntry(); ok {
         // Firefox
         copy(sample.BoxHeader.Type[:], sinf.Frma.DataFormat[:])
      }
   }
   return file1.Append(nil)
}

func (s *senc_test) encode_segment(data []byte) ([]byte, error) {
   fmt.Println(s.segment)
   segment, err := os.ReadFile(s.segment)
   if err != nil {
      return nil, err
   }
   var file1 File
   err = file1.Read(segment)
   if err != nil {
      return nil, err
   }
   track := file1.Moof.Traf
   if senc := track.Senc; senc != nil {
      key, err := hex.DecodeString(s.key)
      if err != nil {
         return nil, err
      }
      for i, data := range file1.Mdat.Data(&track) {
         err := senc.Sample[i].Decrypt(data, key)
         if err != nil {
            return nil, err
         }
      }
   }
   return file1.Append(data)
}

func TestSenc(t *testing.T) {
   for _, test := range senc_tests {
      data, err := test.encode_init()
      if err != nil {
         t.Fatal(err)
      }
      data, err = test.encode_segment(data)
      if err != nil {
         t.Fatal(err)
      }
      err = os.WriteFile(test.dst, data, os.ModePerm)
      if err != nil {
         t.Fatal(err)
      }
   }
}

type senc_test struct {
   init    string
   segment string
   key     string
   dst     string
}
