package sofia

import (
   "os"
   "testing"
)

func TestFile(t *testing.T) {
   for _, test := range tests {
      func() {
         src, err := os.Open(test.init)
         if err != nil {
            t.Fatal(err)
         }
         defer src.Close()
         var dst File
         if err := dst.Read(src); err != nil {
            t.Fatal(err)
         }
      }()
   }
}
