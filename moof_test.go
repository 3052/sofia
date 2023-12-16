package sofia

import (
   "fmt"
   "os"
   "testing"
)

func Test_Moof(t *testing.T) {
   f, err := os.Open("index_video_5_0_1.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer f.Close()
   var m MovieFragmentBox
   if err := m.Decode(f); err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", m)
}
