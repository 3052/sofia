package pssh

import (
   "bytes"
   "encoding/base64"
   "fmt"
   "reflect"
   "testing"
)

const cenc_pssh = "AAAAcHBzc2gAAAAA7e+LqXnWSs6jyCfc1R0h7QAAAFAIARIQmlNKHxLWjhojWfOHEP3bZRoFd3Vha2kiLTlhNTM0YTFmMTJkNjhlMWEyMzU5ZjM4NzEwZmRkYjY1LW1jLTAtMTQ3LTAtMEjj3JWbBg=="

func TestPssh(t *testing.T) {
   read := func() *bytes.Reader {
      b, err := base64.StdEncoding.DecodeString(cenc_pssh)
      if err != nil {
         t.Fatal(err)
      }
      return bytes.NewReader(b)
   }()
   var pssh Box
   err := pssh.BoxHeader.Read(read)
   if err != nil {
      t.Fatal(err)
   }
   err = pssh.Read(read)
   if err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", pssh)
}
