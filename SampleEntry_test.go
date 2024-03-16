package sofia

import (
   "fmt"
   "os"
   "testing"
)

func TestSampleEntry(t *testing.T) {
   src, err := os.Open("testdata/mubi-stpp/textstream_eng=1000-1174000.dash")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var dst File
   if err := dst.Decode(src); err != nil {
      t.Fatal(err)
   }
   if len(dst.MediaData.Data) != 1 {
      t.Fatal("MediaDataBox.Data")
   }
   data := dst.MediaData.Data[0]
   fmt.Println(string(data))
}
