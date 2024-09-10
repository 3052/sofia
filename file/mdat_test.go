package file

import (
   "os"
   "testing"
)

var mdat_tests = []string{
   "../testdata/max-ec-3/segment-1024.m4s",
   "../testdata/max-ec-3/segment-512.m4s",
}

func TestMdat(t *testing.T) {
   for _, test := range mdat_tests {
      func() {
         src, err := os.Open(test)
         if err != nil {
            t.Fatal(err)
         }
         defer src.Close()
         var value File
         err = value.Read(src)
         if err != nil {
            t.Fatal(err)
         }
         value.MediaData.Data(value.MovieFragment.TrackFragment)
      }()
   }
}
