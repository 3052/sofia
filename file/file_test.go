package file

import (
   "41.neocities.org/sofia"
   "os"
   "testing"
)

const file_test = "../testdata/cineMember/knivesout-video-drm-video_eng=492447-0.dash"

func TestFile(t *testing.T) {
   sofia.Debug.SetOutput(os.Stderr)
   data, err := os.ReadFile(file_test)
   if err != nil {
      t.Fatal(err)
   }
   var fileVar File
   err = fileVar.Read(data)
   if err != nil {
      t.Fatal(err)
   }
}
