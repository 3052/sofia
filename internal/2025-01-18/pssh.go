package main

import (
   "41.neocities.org/protobuf"
   "41.neocities.org/sofia/pssh"
   "encoding/base64"
   "encoding/json"
   "fmt"
   "os"
)

const cenc_pssh = "AAAASnBzc2gAAAAA7e+LqXnWSs6jyCfc1R0h7QAAACoSEAAAAABnRkDZbsE0kQf/Je4SEAAAAABnRkDZbsE0kQf/Je9I49yVmwY="

func main() {
   data, err := base64.StdEncoding.DecodeString(cenc_pssh)
   if err != nil {
      panic(err)
   }
   var box pssh.Box
   n, err := box.BoxHeader.Decode(data)
   if err != nil {
      panic(err)
   }
   err = box.Read(data[n:])
   if err != nil {
      panic(err)
   }
   encode := json.NewEncoder(os.Stdout)
   encode.SetIndent("", " ")
   err = encode.Encode(box)
   if err != nil {
      panic(err)
   }
   message := protobuf.Message{}
   err = message.Unmarshal(box.Data)
   if err != nil {
      panic(err)
   }
   fmt.Printf("%#v\n", message)
}
