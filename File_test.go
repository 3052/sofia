package sofia

import (
   "os"
   "testing"
)

func TestCine(t *testing.T) {
   in, err := os.Open("testdata/cine-member/video_eng=110000.dash")
   if err != nil {
      t.Fatal(err)
   }
   defer in.Close()
   var out File
   err = out.Read(in)
   if err != nil {
      t.Fatal(err)
   }
}
