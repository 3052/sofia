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
   "isml",
   "faststart",
   "omit_tfhd_offset",
   "disable_chpl",
   "default_base_moof",
   "dash",
   "cmaf",
   "frag_discont",
   "delay_moov",
   "global_sidx",
   "skip_sidx",
   "write_colr",
   "prefer_icc",
   "write_gama",
   "use_metadata_tags",
   "skip_trailer",
   "negative_cts_offsets",
}

func main() {
   for _, flag := range movflags {
      arg := []string{
         "-i", "BCRhBaFqtf0.mp4",
         "-c", "copy",
         "-frag_size", "6M",
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
