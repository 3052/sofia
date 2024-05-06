package sofia

import (
   "fmt"
   "os"
   "testing"
)

func TestFile(t *testing.T) {
   from, err := os.Open("testdata/draken/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer from.Close()
   var to File
   err = to.Read(from)
   if err != nil {
      t.Fatal(err)
   }
   fmt.Printf("%+v\n", to.Movie.Protection)
}
