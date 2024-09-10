package file

import (
   "os"
   "testing"
)

var media_data_tests = []string{
   "testdata/max-ec-3/segment-1024.m4s",
   "testdata/max-ec-3/segment-512.m4s",
}

func TestMediaData(t *testing.T) {
   for _, test := range media_data_tests {
      func() {
         in, err := os.Open(test)
         if err != nil {
            t.Fatal(err)
         }
         defer in.Close()
         var out File
         err = out.Read(in)
         if err != nil {
            t.Fatal(err)
         }
         out.MediaData.Data(out.MovieFragment.TrackFragment)
      }()
   }
}
