package sofia

import (
   "os"
   "testing"
)

var media_data_tests = []string{
   "testdata/criterion-avc1/sid=1.mp4",
   "testdata/max-ec-3/segment-1024.m4s",
   "testdata/max-ec-3/segment-512.m4s",
}

func TestMediaData(t *testing.T) {
   for _, test := range media_data_tests {
      func() {
         read, err := os.Open(test)
         if err != nil {
            t.Fatal(err)
         }
         defer read.Close()
         var value File
         err = value.Read(read)
         if err != nil {
            t.Fatal(err)
         }
         value.MediaData.Data(value.MovieFragment.TrackFragment)
      }()
   }
}
