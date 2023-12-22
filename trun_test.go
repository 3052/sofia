package sofia

import (
   "encoding/json"
   "fmt"
   "os"
   "testing"
)

func Test_Trun(t *testing.T) {
   media, err := os.Open("testdata/roku-video/index_video_8_0_1.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer media.Close()
   var f File
   if err := f.Decode(media); err != nil {
      t.Fatal(err)
   }
   enc := json.NewEncoder(os.Stdout)
   enc.SetIndent("", " ")
   enc.Encode(f.Moof.Traf.Trun)
   var size uint32
   for _, sample := range f.Moof.Traf.Trun.Samples {
      size += sample.Size
   }
   fmt.Println(size)
}
