package sofia

import (
   "encoding/json"
   "os"
   "testing"
)

func Test_Moof(t *testing.T) {
   f, err := os.Open("index_video_5_0_1.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer f.Close()
   var moof MovieFragmentBox
   if err := moof.Decode(f); err != nil {
      t.Fatal(err)
   }
   enc := json.NewEncoder(os.Stdout)
   enc.SetIndent("", " ")
   enc.Encode(moof)
}
