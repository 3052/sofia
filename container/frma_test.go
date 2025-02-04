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
   var file0 File
   err = file0.Read(data)
   if err != nil {
      t.Fatal(err)
   }
   format := file0.Moov.Trak.Mdia.Minf.Stbl.Stsd.AudioSample.Sinf.Frma
   fmt.Printf("%q\n", format.DataFormat)
}
