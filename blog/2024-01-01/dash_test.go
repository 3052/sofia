package stream

import "testing"

/*
[mfra] size=23190
  [tfra] size=23166 version=1 flags=000000
   - trackID: 1
   - nrEntries: 1218
  [mfro] size=16 version=0 flags=000000
   - parentSize: 23190
*/
func Test_Decrypt(t *testing.T) {
   err := segment_base()
   if err != nil {
      t.Fatal(err)
   }
}
