package sofia

import (
   "os"
   "testing"
)

func Test_Moof(t *testing.T) {
   video, err := os.Open("index_video_5_0_1.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer video.Close()
   var f File
   if err := f.Decode(video); err != nil {
      t.Fatal(err)
   }
}
