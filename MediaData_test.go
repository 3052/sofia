package sofia

import (
   "os"
   "testing"
)

func TestMediaData(t *testing.T) {
   src, err := os.Open("testdata/mubi-stpp\textstream_eng=1000-0.dash")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var dst File
   if err := dst.Decode(src); err != nil {
      t.Fatal(err)
   }
}
