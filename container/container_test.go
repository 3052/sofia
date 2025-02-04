package container

import (
   "os"
   "testing"
)

const file_test = "../testdata/criterion-avc1/0-804.mp4"

func TestFile(t *testing.T) {
   data, err := os.ReadFile(file_test)
   if err != nil {
      t.Fatal(err)
   }
   var file0 File
   err = file0.Read(data)
   if err != nil {
      t.Fatal(err)
   }
}
