package sofia

import (
   "os"
   "testing"
)

func TestFile(t *testing.T) {
   in, err := os.Open("testdata/criterion-mp4a/sid=0.mp4")
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
