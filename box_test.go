package sofia

import (
   "fmt"
   "os"
   "testing"
)

func Test_Box(t *testing.T) {
   f, err := os.Open("index_video_5_0_1.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer f.Close()
   var b Box
   if err := b.Decode(f); err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", b)
}
