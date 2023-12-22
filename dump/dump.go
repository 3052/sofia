package main

import (
   "os"
   "os/exec"
)

func main() {
   if len(os.Args) == 2 {
      input := os.Args[1]
      out, err := exec.Command(
         "mp4tool", "dump", "-full", "tenc,trun", input,
      ).Output()
      if err != nil {
         panic(err)
      }
      os.WriteFile(input + ".txt", out, 0666)
   } else {
      os.Stdout.WriteString("dump [file]\n")
   }
}
