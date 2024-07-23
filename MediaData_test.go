package sofia

import (
   "os"
   "testing"
)

func TestMediaData(t *testing.T) {
   read, err := os.Open("testdata/max-ec-3/segment-1.0001.m4s")
   if err != nil {
      t.Fatal(err)
   }
   defer read.Close()
   var f File
   err = f.Read(read)
   if err != nil {
      t.Fatal(err)
   }
   f.MediaData.Data(f.MovieFragment.TrackFragment.TrackRun)
}
