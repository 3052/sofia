package pssh

import (
   "encoding/base64"
   "fmt"
   "testing"
)

const cenc_pssh = "AAAAcHBzc2gAAAAA7e+LqXnWSs6jyCfc1R0h7QAAAFAIARIQmlNKHxLWjhojWfOHEP3bZRoFd3Vha2kiLTlhNTM0YTFmMTJkNjhlMWEyMzU5ZjM4NzEwZmRkYjY1LW1jLTAtMTQ3LTAtMEjj3JWbBg=="

func TestPssh(t *testing.T) {
   data, err := base64.StdEncoding.DecodeString(cenc_pssh)
   if err != nil {
      t.Fatal(err)
   }
   var pssh Box
   n, err := pssh.BoxHeader.Decode(data)
   if err != nil {
      t.Fatal(err)
   }
   err = pssh.Read(data[n:])
   if err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", pssh)
}
