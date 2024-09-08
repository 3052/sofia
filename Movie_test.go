package sofia

import (
   "os"
   "testing"
)

func TestMovie(t *testing.T) {
   in, err := os.Open("testdata/criterion/sid=0.mp4")
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
