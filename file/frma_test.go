package file

import (
   "fmt"
   "os"
   "testing"
)

func TestFrma(t *testing.T) {
   in, err := os.Open("testdata/hulu-ec-3/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer in.Close()
   var out File
   err = out.Read(in)
   if err != nil {
      t.Fatal(err)
   }
   format := out.
      Movie.
      Track.
      Media.
      MediaInformation.
      SampleTable.
      SampleDescription.
      AudioSample.
      ProtectionScheme.
      OriginalFormat
   fmt.Printf("%q\n", format.DataFormat)
}
