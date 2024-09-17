package main

import (
   "154.pages.dev/sofia/container"
   "encoding/base64"
   "fmt"
   "os"
   "path/filepath"
)

func encode_segment(name string, key []byte) ([]byte, error) {
   buf, err := os.ReadFile(name)
   if err != nil {
      return nil, err
   }
   if key == nil {
      return buf, nil
   }
   var file container.File
   err = file.Read(buf)
   if err != nil {
      return nil, err
   }
   
   return file.Append(nil)
   
   track := file.Moof.Traf
   if senc := track.Senc; senc != nil {
      for i, text := range file.Mdat.Data(&track) {
         err = senc.Sample[i].DecryptCenc(text, key)
         if err != nil {
            return nil, err
         }
      }
   }
   return file.Append(nil)
}

const raw_key = "IcxMAcVIDxupcA55ivgAcw=="

func main() {
   file, err := os.Create("out.mp4")
   if err != nil {
      panic(err)
   }
   defer file.Close()
   buf, err := encode_segment("../persona/init-0.mp4", nil)
   if err != nil {
      panic(err)
   }
   _, err = file.Write(buf)
   if err != nil {
      panic(err)
   }
   matches, err := filepath.Glob("../persona/segment-*.mp4")
   if err != nil {
      panic(err)
   }
   key, err := base64.StdEncoding.DecodeString(raw_key)
   if err != nil {
      panic(err)
   }
   for i, match := range matches {
      fmt.Println(len(matches)-i)
      buf, err = encode_segment(match, key)
      if err != nil {
         panic(err)
      }
      _, err = file.Write(buf)
      if err != nil {
         panic(err)
      }
      break
   }
}
