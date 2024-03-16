package text

import (
   "fmt"
   "os"
   "testing"
)

const name = "../../testdata/mubi-stpp/textstream_eng=1000-1174000.dash"

func TestText(t *testing.T) {
   file, err := os.Open(name)
   if err != nil {
      t.Fatal(err)
   }
   defer file.Close()
   var mark Markup
   if err := mark.Decode(file); err != nil {
      t.Fatal(err)
   }
   fmt.Println(WebVtt)
   for _, p := range mark.Body.Div.P {
      fmt.Println(p)
   }
}
