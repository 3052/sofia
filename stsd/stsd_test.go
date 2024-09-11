package stsd

import (
   "fmt"
   "reflect"
   "testing"
)

var size_test Box

func TestSize(t *testing.T) {
   size := reflect.TypeOf(&struct{}{}).Size()
   if reflect.TypeOf(size_test).Size() > size {
      fmt.Printf("*%T\n", size_test)
   } else {
      fmt.Printf("%T\n", size_test)
   }
}
