package main

import (
   "41.neocities.org/sofia/container"
   "41.neocities.org/sofia/sidx"
   "encoding/base64"
   "os"
   "path/filepath"
)

func main() {
   file, err := os.Create("out.mp4")
   if err != nil {
      panic(err)
   }
   defer file.Close()
   key, err := base64.StdEncoding.DecodeString(raw_key)
   if err != nil {
      panic(err)
   }
   matches, err := filepath.Glob("../persona/segment-*.mp4")
   if err != nil {
      panic(err)
   }
   // data, err := os.ReadFile("../persona/init-000.mp4")
   if err != nil {
      panic(err)
   }
   // offset, err := file.Write(data)
   if err != nil {
      panic(err)
   }
   var index sidx.Box
   index.EarliestPresentationTime = make([]byte, 4)
   index.FirstOffset = make([]byte, 4)
   index.Reference = make([]sidx.Reference, len(matches))
   // data, err := index.Append(nil)
   if err != nil {
      panic(err)
   }
   // _, err = file.Write(data)
   if err != nil {
      panic(err)
   }
   for i, match := range matches {
      data, err := encode_segment(match, key)
      if err != nil {
         panic(err)
      }
      n, err := file.Write(data)
      if err != nil {
         panic(err)
      }
      index.Reference[i].SetSize(uint32(n))
   }
   copy(index.BoxHeader.Type[:], "sidx")
   index.ReferenceCount = uint16(len(matches))
   index.BoxHeader.Size = uint32(index.GetSize())
   // data, err = index.Append(nil)
   if err != nil {
      panic(err)
   }
   // _, err = file.WriteAt(data, int64(offset))
   if err != nil {
      panic(err)
   }
}

func encode_segment(name string, key []byte) ([]byte, error) {
   data, err := os.ReadFile(name)
   if err != nil {
      return nil, err
   }
   var file container.File
   err = file.Read(data)
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

const raw_key = "IcxMAcVIDxupcA55ivgAcw=="
