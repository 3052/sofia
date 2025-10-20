package mp4parser

import (
   "bytes"
   "errors"
   "fmt"
   "log"
   "os"
   "testing"
)

var parser_tests = []struct {
   init    string
   segment string
}{
   {
      "amc-avc1/init.m4f",
      "amc-avc1/segment0.m4f",
   },
   {
      "amc-mp4a/init.m4f",
      "amc-mp4a/segment0.m4f",
   },
   {
      "cineMember-avc1/video_eng=108536.dash",
      "cineMember-avc1/video_eng=108536-0.dash",
   },
   {
      "hboMax-dvh1/init.mp4",
      "hboMax-dvh1/segment-1.0001.m4s",
   },
   {
      "hboMax-ec-3/bytes=0-19985.mp4",
      "hboMax-ec-3/bytes=19986-149146.mp4",
   },
   {
      "hboMax-hvc1/init.mp4",
      "hboMax-hvc1/segment-1.0001.m4s",
   },
   {
      "mubi-avc1/video=300168.dash",
      "mubi-avc1/video=300168-0.dash",
   },
   {
      "mubi-mp4a/audio_eng=268840.dash",
      "mubi-mp4a/audio_eng=268840-0.dash",
   },
   {
      "nbc-avc1/_227156876_5.mp4",
      "nbc-avc1/_227156876_5_0.mp4",
   },
   {
      "nbc-mp4a/_227156876_6_1.mp4",
      "nbc-mp4a/_227156876_6_1_0.mp4",
   },
   {
      "paramount-avc1/0-17641.mp4",
      "paramount-avc1/17642-196772.mp4",
   },
   {
      "paramount-mp4a/init.m4v",
      "paramount-mp4a/seg_1.m4s",
   },
   {
      "plex-avc1/video_init.mp4",
      "plex-avc1/video_1.m4s",
   },
   {
      "roku-avc1/index_video_8_0_init.mp4",
      "roku-avc1/index_video_8_0_1.mp4",
   },
   {
      "roku-mp4a/index_audio_2_0_init.mp4",
      "roku-mp4a/index_audio_2_0_1.mp4",
   },
   {
      "criterion-avc1/0-804.mp4",
      "criterion-avc1/13845-168166.mp4",
   },
   {
      "criterion-mp4a/0-677.mp4",
      "criterion-mp4a/13730-159186.mp4",
   },
   {
      "",
      "hulu-avc1/pts_0.mp4",
   },
   {
      "rtbf/vod-idx-video=4000000.dash",
      "",
   },
}

func do_parse(name string) error {
   name = folder + name
   log.Print(name)
   sampleFMP4, err := os.ReadFile(name)
   if err != nil {
      return err
   }
   parser := NewParser(sampleFMP4)
   var (
      parsedBoxes     []*Box
      totalParsedSize uint64
   )
   for parser.HasMore() {
      box, err := parser.ParseNextBox()
      if err != nil {
         return fmt.Errorf("Failed to parse box: %v", err)
      }
      if box == nil {
         break
      }
      parsedBoxes = append(parsedBoxes, box)
      totalParsedSize += box.Header.Size
   }
   if totalParsedSize != uint64(len(sampleFMP4)) {
      return fmt.Errorf(
         "did not consume the entire file: got %d bytes, want %d bytes",
         totalParsedSize, len(sampleFMP4),
      )
   }
   formattedBuffer := new(bytes.Buffer)
   for _, box := range parsedBoxes {
      formattedBytes, err := box.Format()
      if err != nil {
         return fmt.Errorf("format box of type '%s': %v", box.Header.Type, err)
      }
      formattedBuffer.Write(formattedBytes)
   }
   formattedData := formattedBuffer.Bytes()
   if len(sampleFMP4) != len(formattedData) {
      return fmt.Errorf(
         "original is %d bytes, formatted is %d bytes",
         len(sampleFMP4), len(formattedData),
      )
   }
   if !bytes.Equal(sampleFMP4, formattedData) {
      return errors.New("formatted data does not match original data")
   }
   return nil
}

func TestParser(t *testing.T) {
   for _, test := range parser_tests {
      err := do_parse(test.init)
      if err != nil {
         t.Fatal(err)
      }
      err = do_parse(test.segment)
      if err != nil {
         t.Fatal(err)
      }
   }
}

const folder = "../../testdata/"
