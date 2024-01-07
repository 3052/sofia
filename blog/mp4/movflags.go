package main

import (
   "fmt"
   "os"
   "os/exec"
)

var movflags = []string{
   "",
   "rtphint",
   "empty_moov",
   "frag_keyframe",
   "frag_every_frame",
   "separate_moof",
   "frag_custom",
   "cmaf",
   "dash",
   "default_base_moof",
   "delay_moov",
   "disable_chpl",
   "faststart",
   "frag_discont",
   "global_sidx",
   "negative_cts_offsets",
   "omit_tfhd_offset",
   "prefer_icc",
   "skip_sidx",
   "skip_trailer",
   "use_metadata_tags",
   "write_colr",
   "write_gama",
}

func main() {
   for _, flag := range movflags {
      arg := []string{
         "-i", "in.mp4",
         "-c", "copy",
      }
      if flag != "" {
         arg = append(arg, "-movflags", flag)
      }
      arg = append(arg, flag + ".mp4")
      cmd := exec.Command("ffmpeg", arg...)
      fmt.Println(cmd.Args)
      err := cmd.Run()
      if err != nil {
         panic(err)
      }
      func() {
         file, err := os.Create(flag + ".txt")
         if err != nil {
            panic(err)
         }
         defer file.Close()
         cmd := exec.Command("mp4ff-info", flag + ".mp4")
         cmd.Stdout = file
         fmt.Println(cmd.Args)
         if err := cmd.Run(); err != nil {
            panic(err)
         }
      }()
   }
}
