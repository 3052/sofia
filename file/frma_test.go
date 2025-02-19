package file

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
   var file1 File
   err = file1.Read(data)
   if err != nil {
      t.Fatal(err)
   }
   format := file1.Moov.Trak.Mdia.Minf.Stbl.Stsd.AudioSample.Sinf.Frma
   fmt.Printf("%q\n", format.DataFormat)
}
