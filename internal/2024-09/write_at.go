package main

import "os"

func main() {
   file, err := os.Create("hello.txt")
   if err != nil {
      panic(err)
   }
   defer file.Close()
   _, err = file.WriteString("alfa bravo charlie\n")
   if err != nil {
      panic(err)
   }
   _, err = file.WriteAt([]byte("BRAVO"), 5)
   if err != nil {
      panic(err)
   }
}
