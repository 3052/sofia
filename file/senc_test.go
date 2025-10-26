package file

import (
   "encoding/hex"
   "log"
   "os"
   "testing"
)

func (s *senc_test) encode_init() ([]byte, error) {
   log.Print(s.initial)
   data, err := os.ReadFile(folder + s.initial)
   if err != nil {
      return nil, err
   }
   var fileVar File
   err = fileVar.Read(data)
   if err != nil {
      return nil, err
   }
   for _, pssh := range fileVar.Moov.Pssh {
      copy(pssh.BoxHeader.Type[:], "free") // Firefox
   }
   description := fileVar.Moov.Trak.Mdia.Minf.Stbl.Stsd
   if sinf, ok := description.Sinf(); ok {
      // Firefox
      copy(sinf.BoxHeader.Type[:], "free")
      if sample, ok := description.SampleEntry(); ok {
         // Firefox
         copy(sample.BoxHeader.Type[:], sinf.Frma.DataFormat[:])
      }
   }
   return fileVar.Append(nil)
}

func (s *senc_test) encode_segment(data []byte) ([]byte, error) {
   log.Print(s.segment)
   segment, err := os.ReadFile(folder + s.segment)
   if err != nil {
      return nil, err
   }
   var fileVar File
   err = fileVar.Read(segment)
   if err != nil {
      return nil, err
   }
   track := fileVar.Moof.Traf
   if senc := track.Senc; senc != nil {
      key, err := hex.DecodeString(s.key)
      if err != nil {
         return nil, err
      }
      for i, data := range fileVar.Mdat.Data(&track) {
         err := senc.Sample[i].Decrypt(data, key)
         if err != nil {
            return nil, err
         }
      }
   }
   return fileVar.Append(data)
}

const folder = "../testdata/"

type senc_test struct {
   initial string
   key     string
   out     string
   segment string
}

var senc_tests = []senc_test{
   //{
   //   initial: "criterion-avc1/0-804.mp4",
   //   key:     "377772323b0f45efb2c53c603749d834",
   //   out:     "criterion-avc1.mp4",
   //   segment: "criterion-avc1/13845-168166.mp4",
   //},
   //{
   //   initial: "hboMax-dvh1/0-862.mp4",
   //   key:     "8ea21645755811ecb84b1f7c39bbbff3",
   //   out:     "hboMax-dvh1.mp4",
   //   segment: "hboMax-dvh1/19579-78380.mp4",
   //},
   //{
   //   initial: "hboMax-ec-3/0-657.mp4",
   //   key:     "acaec99945a3615c9ef7b1b04727022a",
   //   out:     "hboMax-ec-3.mp4",
   //   segment: "hboMax-ec-3/28710-157870.mp4",
   //},
   //{
   //   initial: "hboMax-hvc1/0-834.mp4",
   //   key:     "a269d5aebc114fd167c380f801437f49",
   //   out:     "hboMax-hvc1.mp4",
   //   segment: "hboMax-hvc1/19551-35438.mp4",
   //},
   //{
   //   initial: "hulu-avc1/map.mp4",
   //   key:     "33a7ef13ee16fa6a3d1467c0cc59a84f",
   //   out:     "hulu-avc1.mp4",
   //   segment: "hulu-avc1/pts_0.mp4",
   //},
   //{
   //   initial: "paramount-mp4a/init.m4v",
   //   key:     "d98277ff6d7406ec398b49bbd52937d4",
   //   out:     "paramount-mp4a.mp4",
   //   segment: "paramount-mp4a/seg_1.m4s",
   //},
   {
      initial: "roku-avc1/index_video_8_0_init.mp4",
      key:     "1ba08384626f9523e37b9db17f44da2b",
      out:     "roku-avc1.mp4",
      segment: "roku-avc1/index_video_8_0_1.mp4",
   },
   //{
   //   initial: "rtbf-avc1/vod-idx-3-video=300000.dash",
   //   key:     "553b091b257584d3938c35dd202531f8",
   //   out:     "rtbf-avc1.mp4",
   //   segment: "rtbf-avc1/vod-idx-3-video=300000-0.dash",
   //},
   //{
   //   initial: "tubi-avc1/0-1683.mp4",
   //   key:     "8109222ffe94120d61f887d40d0257ed",
   //   out:     "tubi-avc1.mp4",
   //   segment: "tubi-avc1/16524-27006.mp4",
   //},
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
