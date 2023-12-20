package sofia

import (
   "bytes"
   "encoding/binary"
   "fmt"
)

func UUID(uuid_box []byte) {
   r := bytes.NewReader(uuid_box)
   var usertype [16]byte // a2394f52-5a9b-4f14-a244-6c427c648df4
   r.Read(usertype[:])
   // fullBoxHeader
   var fullBoxHeader struct {
      Version int8
      Flags   [3]byte
   }
   err := binary.Read(r, nil, &fullBoxHeader)
   if err != nil {
      panic(err)
   }
   // sample_count
   var sample_count uint32
   err = binary.Read(r, binary.BigEndian, &sample_count)
   if err != nil {
      panic(err)
   }
   for sample_count >= 1 {
      var a struct {
         InitializationVector [8]byte
         NumberOfEntries      uint16
      }
      err := binary.Read(r, binary.BigEndian, &a)
      if err != nil {
         panic(err)
      }
      for a.NumberOfEntries >= 1 {
         var b struct {
            BytesOfClearData     uint16
            BytesOfEncryptedData uint32
         }
         err := binary.Read(r, binary.BigEndian, &b)
         if err != nil {
            panic(err)
         }
         fmt.Printf("%+v\n", b)
         a.NumberOfEntries--
      }
      sample_count--
   }
}
