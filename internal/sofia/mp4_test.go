package mp4

import (
   "bytes"
   "errors"
   "log"
   "os"
   "path/filepath"
   "testing"
)

// parseAndEncodeTopLevelBox is a dispatcher for the specific box parsers.
// It parses a single box from the start of the data slice, re-encodes it,
// and returns the encoded data and the size of the original box consumed.
func parseAndEncodeTopLevelBox(data []byte) ([]byte, int, error) {
   if len(data) < 8 {
      return nil, 0, errors.New("not enough data for box header")
   }
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return nil, 0, err
   }
   boxType := string(header.Type[:])

   // Ensure the slice is large enough for the declared box size.
   // Note: A size of 0 means the box extends to the end of the file.
   // This test implementation does not handle that case and assumes fixed sizes.
   boxSize := int(header.Size)
   if len(data) < boxSize {
      return nil, 0, errors.New("box size is larger than available data")
   }

   // This is the raw data for the current box.
   originalBoxData := data[:boxSize]

   var encodedBoxData []byte
   var parseErr error

   // A map of parsers for the boxes you care about.
   parsers := map[string]func([]byte) ([]byte, error){
      "enca": func(d []byte) ([]byte, error) {
         b, e := ParseEnca(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "encv": func(d []byte) ([]byte, error) {
         b, e := ParseEncv(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "frma": func(d []byte) ([]byte, error) {
         b, e := ParseFrma(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "mdat": func(d []byte) ([]byte, error) {
         b, e := ParseMdat(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "mdhd": func(d []byte) ([]byte, error) {
         b, e := ParseMdhd(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "mdia": func(d []byte) ([]byte, error) {
         b, e := ParseMdia(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "minf": func(d []byte) ([]byte, error) {
         b, e := ParseMinf(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "moof": func(d []byte) ([]byte, error) {
         b, e := ParseMoof(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "moov": func(d []byte) ([]byte, error) {
         b, e := ParseMoov(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "pssh": func(d []byte) ([]byte, error) {
         b, e := ParsePssh(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "schi": func(d []byte) ([]byte, error) {
         b, e := ParseSchi(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "senc": func(d []byte) ([]byte, error) {
         b, e := ParseSenc(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "sidx": func(d []byte) ([]byte, error) {
         b, e := ParseSidx(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "sinf": func(d []byte) ([]byte, error) {
         b, e := ParseSinf(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "stbl": func(d []byte) ([]byte, error) {
         b, e := ParseStbl(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "stsd": func(d []byte) ([]byte, error) {
         b, e := ParseStsd(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "tenc": func(d []byte) ([]byte, error) {
         b, e := ParseTenc(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "tfhd": func(d []byte) ([]byte, error) {
         b, e := ParseTfhd(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "traf": func(d []byte) ([]byte, error) {
         b, e := ParseTraf(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "trak": func(d []byte) ([]byte, error) {
         b, e := ParseTrak(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
      "trun": func(d []byte) ([]byte, error) {
         b, e := ParseTrun(d)
         if e != nil {
            return nil, e
         }
         return b.Encode(), nil
      },
   }

   if parser, found := parsers[boxType]; found {
      encodedBoxData, parseErr = parser(originalBoxData)
   } else {
      // Per your request, other boxes remain unparsed. For a byte-accurate
      // round trip, we must return their original data unmodified.
      encodedBoxData = originalBoxData
   }

   if parseErr != nil {
      return nil, 0, parseErr
   }

   return encodedBoxData, boxSize, nil
}

// TestRoundTrip loops through specified files and ensures that parsing and
// then encoding the data results in the exact same byte sequence.
func TestRoundTrip(t *testing.T) {
   // The user must place the test files in a 'testdata' subdirectory
   // relative to the module root.
   testFiles := []string{
      "testdata/criterion-avc1/0-804.mp4",
      "testdata/criterion-avc1/13845-168166.mp4",
      "testdata/hboMax-dvh1/0-862.mp4",
      "testdata/hboMax-dvh1/19579-78380.mp4",
      "testdata/hulu-avc1/map.mp4",
      "testdata/hulu-avc1/pts_0.mp4",
      "testdata/paramount-mp4a/init.m4v",
      "testdata/paramount-mp4a/seg_1.m4s",
      "testdata/roku-avc1/index_video_8_0_1.mp4",
      "testdata/tubi-avc1/0-1683.mp4",
      "testdata/tubi-avc1/16524-27006.mp4",
   }

   for _, filePath := range testFiles {
      log.Print(filePath)
      // Use a subtest for each file for clearer test output.
      t.Run(filepath.Base(filePath), func(t *testing.T) {
         originalData, err := os.ReadFile("../../" + filePath)
         if err != nil {
            // Skip the test if the file doesn't exist, but notify the user
            // so they know they need to provide it.
            t.Skipf("test file not found, skipping: %s", filePath)
            return
         }

         if len(originalData) == 0 {
            t.Logf("test file is empty: %s", filePath)
            return
         }

         var encodedData []byte
         offset := 0
         for offset < len(originalData) {
            encodedBox, size, err := parseAndEncodeTopLevelBox(originalData[offset:])
            if err != nil {
               t.Fatalf("failed to parse/encode box in file %s at offset %d: %v", filePath, offset, err)
            }
            if size == 0 {
               t.Fatalf("parsed box with size 0, cannot advance offset")
            }

            encodedData = append(encodedData, encodedBox...)
            offset += size
         }

         if !bytes.Equal(originalData, encodedData) {
            t.Errorf("Round trip failed for %s. Original and encoded data do not match.", filePath)

            // For debugging, you can uncomment these lines to write the output files
            // and inspect the differences with a hex editor.
            // debugDir := "debug_output"
            // os.Mkdir(debugDir, 0755)
            // baseName := filepath.Base(filePath)
            // os.WriteFile(filepath.Join(debugDir, "original_"+baseName), originalData, 0644)
            // os.WriteFile(filepath.Join(debugDir, "encoded_"+baseName), encodedData, 0644)
         }
      })
   }
}
