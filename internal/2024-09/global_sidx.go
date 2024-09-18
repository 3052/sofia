package main

import (
   "154.pages.dev/sofia/container"
   "encoding/base64"
   "fmt"
   "os"
   "path/filepath"
)

const raw_key = "IcxMAcVIDxupcA55ivgAcw=="

func main() {
   file, err := os.Create("out.mp4")
   if err != nil {
      panic(err)
   }
   defer file.Close()
   buf, err := os.ReadFile("../persona/init-000.mp4")
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
   for _, match := range matches {
      fmt.Println(match)
      buf, err = encode_segment(match, key)
      if err != nil {
         panic(err)
      }
      _, err = file.Write(buf)
      if err != nil {
         panic(err)
      }
   }
}

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
   track := file.Moof.Traf
   for _, box := range track.Box {
      if box.BoxHeader.Type.String() == "saio" {
         copy(box.BoxHeader.Type[:], "free") // mp4ff-info
      }
   }
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