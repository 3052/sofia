package file

import (
   "fmt"
   "os"
   "testing"
)

func TestTenc(t *testing.T) {
   src, err := os.Open("../testdata/amc-avc1/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var value File
   err = value.Read(src)
   if err != nil {
      t.Fatal(err)
   }
   protect, ok := value.
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
