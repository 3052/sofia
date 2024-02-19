package main

import (
   "fmt"
   "os"
   "os/exec"
)

func main() {
   if len(os.Args) == 2 {
      input := os.Args[1]
      cmd := exec.Command("mp4ff-info", input)
      fmt.Println(cmd.Args)
      out, err := cmd.Output()
      if err != nil {
         panic(err)
      }
      os.WriteFile(input+".txt", out, 0666)
   } else {
      fmt.Println("dump [file]")
   }
}
