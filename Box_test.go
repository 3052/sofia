package sofia

import (
   "fmt"
   "reflect"
   "testing"
)

// 4 sofia.FullBoxHeader
// 4 sofia.Type
// 8 sofia.Subsample
// 8 sofia.box_error

func TestSize(t *testing.T) {
   a := reflect.TypeOf(&struct{}{}).Size()
   for _, test := range size_tests {
      if b := reflect.TypeOf(test).Size(); b <= a {
         fmt.Printf("%v %T\n", b, test)
      }
   }
}

var size_tests = []any{
   AudioSampleEntry{},
   Box{},
   BoxHeader{},
   EncryptionSample{},
   File{},
   FullBoxHeader{},
   Media{},
   MediaData{},
   MediaInformation{},
   Movie{},
   MovieFragment{},
   OriginalFormat{},
   ProtectionSchemeInfo{},
   ProtectionSystemSpecificHeader{},
   Reference{},
   RunSample{},
   SampleDescription{},
   SampleEncryption{},
   SampleEntry{},
   SampleTable{},
   SchemeInformation{},
   SegmentIndex{},
   Subsample{},
   Track{},
   TrackEncryption{},
   TrackFragment{},
   TrackFragmentHeader{},
   TrackRun{},
   Type{},
   UUID{},
   VisualSampleEntry{},
   box_error{},
}
