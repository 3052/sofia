package sofia

import (
   "os"
   "testing"
)

func TestFile(t *testing.T) {
   from, err := os.Open("testdata/joyn/video_eng=351000.dash")
   if err != nil {
      t.Fatal(err)
   }
   defer from.Close()
   var to File
   err = to.Read(from)
   if err != nil {
      t.Fatal(err)
   }
}
