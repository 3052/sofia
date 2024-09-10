package file

import (
   "fmt"
   "os"
   "testing"
)

func TestTenc(t *testing.T) {
   in, err := os.Open("testdata/amc-avc1/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer in.Close()
   var out File
   err = out.Read(in)
   if err != nil {
      t.Fatal(err)
   }
   protect, ok := out.
      Movie.
      Track.
      Media.
      MediaInformation.
      SampleTable.
      SampleDescription.
      Protection()
   if !ok {
      t.Fatal("Protection")
   }
   fmt.Printf("%+v\n", protect.SchemeInformation.TrackEncryption)
}
