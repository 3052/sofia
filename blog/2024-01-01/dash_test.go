package stream

import "testing"

func Test_Stream(t *testing.T) {
   err := segment_base()
   if err != nil {
      t.Fatal(err)
   }
}
