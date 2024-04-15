package main

import (
   "154.pages.dev/sofia"
   "bytes"
   "fmt"
   "os"
)

func main() {
   data, err := os.ReadFile("gn1spumg.mp4")
   if err != nil {
      panic(err)
   }
   var file sofia.File
   slice := sofia.Slice{End: 30057+1}
   err = file.Read(bytes.NewReader(data[:slice.End]))
   if err != nil {
      panic(err)
   }
   for i, reference := range file.SegmentIndex.Reference {
      slice.Add(reference)
      fmt.Println(slice)
      var file sofia.File
      err := file.Read(bytes.NewReader(data[slice.Start:slice.End]))
      if err != nil {
         panic(err)
      }
      if file.MovieFragment.TrackFragment.SampleEncryption.Samples == nil {
         fmt.Println(i)
      }
   }
}
