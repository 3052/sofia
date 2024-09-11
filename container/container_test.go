package container

import (
   "fmt"
   "reflect"
   "testing"
)

func TestSize(t *testing.T) {
   size := reflect.TypeOf(&struct{}{}).Size()
   var test File
   if reflect.TypeOf(test).Size() > size {
      fmt.Printf("*%T\n", test)
   } else {
      fmt.Printf("%T\n", test)
   }
}
