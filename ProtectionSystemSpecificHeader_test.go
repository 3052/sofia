package sofia

import (
   "bytes"
   "encoding/base64"
   "fmt"
   "os"
   "testing"
)

func TestM4f(t *testing.T) {
   src, err := os.Open("testdata/amc-avc1/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var dst File
   if err := dst.Read(src); err != nil {
      t.Fatal(err)
   }
   for _, protect := range dst.Movie.Protection {
      fmt.Println(protect.SystemId)
   }
}

const cenc_pssh = "AAAAVnBzc2gAAAAA7e+LqXnWSs6jyCfc1R0h7QAAADYIARIQXn02m57KRCakPhWnbwndfhoNd2lkZXZpbmVfdGVzdCIIMTIzNDU2NzgyB2RlZmF1bHQ="

func TestMpd(t *testing.T) {
   r := func() *bytes.Reader {
      b, err := base64.StdEncoding.DecodeString(cenc_pssh)
      if err != nil {
         t.Fatal(err)
      }
      return bytes.NewReader(b)
   }()
   var protect ProtectionSystemSpecificHeader
   err := protect.BoxHeader.read(r)
   if err != nil {
      t.Fatal(err)
   }
   if err := protect.read(r); err != nil {
      t.Fatal(err)
   }
   fmt.Println(protect.SystemId)
}
