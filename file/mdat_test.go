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
      data, err := os.ReadFile(test)
      if err != nil {
         t.Fatal(err)
      }
      var file1 File
      err = file1.Read(data)
      if err != nil {
         t.Fatal(err)
      }
      file1.Mdat.Data(&file1.Moof.Traf)
   }
}
