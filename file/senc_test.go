package file

import (
   "encoding/hex"
   "log"
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
      "hboMax-dvh1/0-902.mp4",
      "hboMax-dvh1/903-28882.mp4",
      "ee0d569c019057569eaf28b988c206f6",
      "hboMax-dvh1.mp4",
   },
   {
      "hboMax-ec-3/0-657.mp4",
      "hboMax-ec-3/658-28709.mp4",
      "acaec99945a3615c9ef7b1b04727022a",
      "hboMax-ec-3.mp4",
   },
   {
      "hboMax-hvc1/0-793.mp4",
      "hboMax-hvc1/794-28773.mp4",
      "bd691b57ac7c0620482c724b953a8e87",
      "hboMax-hvc1.mp4",
   },
   {
      "hulu-avc1/map.mp4",
      "hulu-avc1/pts_0.mp4",
      "33a7ef13ee16fa6a3d1467c0cc59a84f",
      "hulu-avc1.mp4",
   },
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
   {
      "rtbf-avc1/vod-idx-3-video=300000.dash",
      "rtbf-avc1/vod-idx-3-video=300000-0.dash",
      "553b091b257584d3938c35dd202531f8",
      "rtbf-avc1.mp4",
   },
   {
      "tubi-avc1/0-1683.mp4",
      "tubi-avc1/1684-16523.mp4",
      "8109222ffe94120d61f887d40d0257ed",
      "tubi-avc1.mp4",
   },
}

func (s *senc_test) encode_init() ([]byte, error) {
   data, err := os.ReadFile(folder + s.initial)
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
   log.Print(folder + s.segment)
   segment, err := os.ReadFile(folder + s.segment)
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
      err = os.WriteFile(test.out, data, os.ModePerm)
      if err != nil {
         t.Fatal(err)
      }
   }
}

type senc_test struct {
   initial string
   segment string
   key     string
   out     string
}
