package mp4parser

import (
   "bytes"
   "os"
   "testing"
)

var names = []string{
   `..\..\testdata\amc-avc1\init.m4f`,
   `..\..\testdata\amc-avc1\segment0.m4f`,
   `..\..\testdata\amc-mp4a\init.m4f`,
   `..\..\testdata\amc-mp4a\segment0.m4f`,
   `..\..\testdata\cineMember-avc1\video_eng=108536-0.dash`,
   `..\..\testdata\cineMember-avc1\video_eng=108536.dash`,
   `..\..\testdata\criterion-avc1\0-804.mp4`,
   `..\..\testdata\criterion-mp4a\sid=0.mp4`,
   `..\..\testdata\ctv\init.mp4`,
   `..\..\testdata\draken\init.mp4`,
   `..\..\testdata\hulu-avc1\pts_0.mp4`,
   `..\..\testdata\hulu-ec-3\init.mp4`,
   `..\..\testdata\hulu-ec-3\segment-1.0001.m4s`,
   `..\..\testdata\hulu-hev1\init.mp4`,
   `..\..\testdata\hulu-hev1\segment-1.0001.m4s`,
   `..\..\testdata\hulu-mp4a\init.mp4`,
   `..\..\testdata\hulu-mp4a\segment-1.0001.m4s`,
   `..\..\testdata\max-dvh1\init.mp4`,
   `..\..\testdata\max-dvh1\segment-1.0001.m4s`,
   `..\..\testdata\max-ec-3\bytes=0-19985.mp4`,
   `..\..\testdata\max-ec-3\bytes=19986-149146.mp4`,
   `..\..\testdata\max-hvc1\init.mp4`,
   `..\..\testdata\max-hvc1\segment-1.0001.m4s`,
   `..\..\testdata\mubi-avc1\video=300168-0.dash`,
   `..\..\testdata\mubi-avc1\video=300168.dash`,
   `..\..\testdata\mubi-mp4a\audio_eng=268840-0.dash`,
   `..\..\testdata\mubi-mp4a\audio_eng=268840.dash`,
   `..\..\testdata\nbc-avc1\_227156876_5.mp4`,
   `..\..\testdata\nbc-avc1\_227156876_5_0.mp4`,
   `..\..\testdata\nbc-mp4a\_227156876_6_1.mp4`,
   `..\..\testdata\nbc-mp4a\_227156876_6_1_0.mp4`,
   `..\..\testdata\paramount-avc1\0-17641.mp4`,
   `..\..\testdata\paramount-avc1\17642-196772.mp4`,
   `..\..\testdata\paramount-mp4a\init.m4v`,
   `..\..\testdata\paramount-mp4a\seg_1.m4s`,
   `..\..\testdata\plex-avc1\video_1.m4s`,
   `..\..\testdata\plex-avc1\video_init.mp4`,
   `..\..\testdata\roku-avc1\index_video_8_0_1.mp4`,
   `..\..\testdata\roku-avc1\index_video_8_0_init.mp4`,
   `..\..\testdata\roku-mp4a\index_audio_2_0_1.mp4`,
   `..\..\testdata\roku-mp4a\index_audio_2_0_init.mp4`,
   `..\..\testdata\rtbf\vod-idx-video=4000000.dash`,
   `..\..\testdata\tubi-avc1\0-30057.mp4`,
   `..\..\testdata\tubi-avc1\30058-111481.mp4`,
   `..\..\testdata\tubi-mp4a\0-1547.mp4`,
}

func TestRoundtrip(t *testing.T) {
   for _, name := range names {
      t.Log(name)
      sampleFMP4, err := os.ReadFile(name)
      if err != nil {
         t.Fatal(err)
      }
      parser := NewParser(sampleFMP4)
      var (
         parsedBoxes []*Box
         totalParsedSize uint64
      )
      for parser.HasMore() {
         box, err := parser.ParseNextBox()
         if err != nil {
            t.Fatalf("Failed to parse box: %v", err)
         }
         if box == nil {
            break
         }
         parsedBoxes = append(parsedBoxes, box)
         totalParsedSize += box.Header.Size
      }
      if totalParsedSize != uint64(len(sampleFMP4)) {
         t.Errorf(
            "did not consume the entire file: got %d bytes, want %d bytes",
            totalParsedSize, len(sampleFMP4),
         )
      }
      formattedBuffer := new(bytes.Buffer)
      for _, box := range parsedBoxes {
         formattedBytes, err := box.Format()
         if err != nil {
            t.Fatalf("format box of type '%s': %v", box.Header.Type, err)
         }
         formattedBuffer.Write(formattedBytes)
      }
      formattedData := formattedBuffer.Bytes()
      if len(sampleFMP4) != len(formattedData) {
         t.Fatalf(
            "original is %d bytes, formatted is %d bytes",
            len(sampleFMP4), len(formattedData),
         )
      }
      if !bytes.Equal(sampleFMP4, formattedData) {
         t.Errorf("formatted data does not match original data")
      }
   }
}
