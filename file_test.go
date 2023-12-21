package sofia

import (
   "encoding/json"
   "os"
   "testing"
)

func Test_Moof(t *testing.T) {
   video, err := os.Open("testdata/roku-video/index_video_8_0_1.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer video.Close()
   var f File
   if err := f.Decode(video); err != nil {
      t.Fatal(err)
   }
   enc := json.NewEncoder(os.Stdout)
   enc.SetIndent("", " ")
   enc.Encode(f)
}
