package container

import (
   "fmt"
   "os"
   "testing"
)

func TestFrma(t *testing.T) {
   data, err := os.ReadFile("../testdata/hulu-ec-3/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   var value File
   err = value.Read(data)
   if err != nil {
      t.Fatal(err)
   }
   format := value.Moov.Trak.Mdia.Minf.Stbl.Stsd.AudioSample.Sinf.Frma
   fmt.Printf("%q\n", format.DataFormat)
}
