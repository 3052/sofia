package file

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
   // cineMember-avc1\video_eng=108536-0.dash
   // cineMember-avc1\video_eng=108536.dash
   // 
   // hbomax-dvh1\init.mp4
   // hbomax-dvh1\segment-1.0001.m4s
   // 
   // hbomax-ec-3\bytes=0-19985.mp4
   // hbomax-ec-3\bytes=19986-149146.mp4
   // 
   // hbomax-hvc1\init.mp4
   // hbomax-hvc1\segment-1.0001.m4s
   {
      "../testdata/hulu-avc1/init.mp4",
      "../testdata/hulu-avc1/segment-1.0001.m4s",
      "602a9289bfb9b1995b75ac63f123fc86",
      "hulu-avc1.mp4",
   },
   {
      "../testdata/mubi-mp4a/audio_eng=268840.dash",
      "../testdata/mubi-mp4a/audio_eng=268840-0.dash",
      "2556f746e8db3ee7f66fc22f5a28752a",
      "mubi-mp4a.mp4",
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
