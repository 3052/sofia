// File: senc_test.go
package mp4parser

import (
   "bytes"
   "testing"
)

// buildTestInitSegment programmatically creates a valid init segment byte slice.
func buildTestInitSegment() []byte {
   // Build the box structure from the inside out.
   tencBox := &TencBox{
      RemainingData: []byte{
         0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x08,
         0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
         0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
      },
   }
   schiBox := &SchiBox{Children: []*SchiChildBox{{Tenc: tencBox}}}
   frmaBox := &FrmaBox{DataFormat: []byte{'a', 'v', 'c', '1'}}
   sinfBox := &SinfBox{Children: []*SinfChildBox{{Frma: frmaBox}, {Schi: schiBox}}}
   encvBox := &EncvBox{Type: "encv", Prefix: make([]byte, 78), Children: []*EncvChildBox{{Sinf: sinfBox}}}
   stsdBox := &StsdBox{EntryCount: 1, Children: []*StsdChildBox{{Encv: encvBox}}}
   stblBox := &StblBox{Children: []*StblChildBox{{Stsd: stsdBox}}}
   mdhdBox := &MdhdBox{RemainingData: make([]byte, 24)}
   minfBox := &MinfBox{Children: []*MinfChildBox{{Stbl: stblBox}}}
   mdiaBox := &MdiaBox{Children: []*MdiaChildBox{{Mdhd: mdhdBox}, {Minf: minfBox}}}

   // THIS IS THE CORRECTED PART
   trakBox := &TrakBox{
      Children: []*TrakChildBox{{Mdia: mdiaBox}}, // Correctly use []*TrakChildBox
   }

   moovBox := &MoovBox{Children: []*MoovChildBox{{Trak: trakBox}}}
   topLevelBox := Box{Moov: moovBox}
   finalBytes, err := topLevelBox.Format()
   if err != nil {
      panic("Failed to build test init segment with top-level box: " + err.Error())
   }
   return finalBytes
}

var sampleInitWithTenc = buildTestInitSegment()

var sampleSegmentWithSenc = []byte{
   // moof box
   0x00, 0x00, 0x00, 0x3A, 'm', 'o', 'o', 'f',
   // traf box
   0x00, 0x00, 0x00, 0x32, 't', 'r', 'a', 'f',
   // senc box (raw)
   0x00, 0x00, 0x00, 0x2A, 's', 'e', 'n', 'c',
   0x00, 0x00, 0x00, 0x02,
   0x00, 0x00, 0x00, 0x02,
   0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
   0x00, 0x01,
   0x12, 0x34,
   0x00, 0x00, 0x56, 0x78,
   0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
   0x00, 0x00,
}

func TestParseSencWithContext(t *testing.T) {
   var perSampleIVSize uint8
   initParser := NewParser(sampleInitWithTenc)

   moovBox, err := initParser.ParseNextBox()
   if err != nil || moovBox.Moov == nil {
      t.Fatalf("Failed to parse moov box from init segment: %v", err)
   }

   var tencBox *TencBox
   // Robustly navigate to the 'tenc' box
   for _, moovChild := range moovBox.Moov.Children {
      if moovChild.Trak != nil {
         for _, trakChild := range moovChild.Trak.Children {
            if trakChild.Mdia != nil {
               for _, mdiaChild := range trakChild.Mdia.Children {
                  if mdiaChild.Minf != nil {
                     for _, minfChild := range mdiaChild.Minf.Children {
                        if minfChild.Stbl != nil {
                           for _, stblChild := range minfChild.Stbl.Children {
                              if stblChild.Stsd != nil {
                                 for _, stsdChild := range stblChild.Stsd.Children {
                                    if stsdChild.Encv != nil {
                                       for _, encvChild := range stsdChild.Encv.Children {
                                          if encvChild.Sinf != nil {
                                             for _, sinfChild := range encvChild.Sinf.Children {
                                                if sinfChild.Schi != nil {
                                                   for _, schiChild := range sinfChild.Schi.Children {
                                                      if schiChild.Tenc != nil {
                                                         tencBox = schiChild.Tenc
                                                         break
                                                      }
                                                   }
                                                }
                                             }
                                          }
                                       }
                                    }
                                 }
                              }
                           }
                        }
                     }
                  }
               }
            }
         }
      }
   }

   if tencBox == nil {
      t.Fatal("Failed to find 'tenc' box in the init segment")
   }

   if len(tencBox.RemainingData) < 8 {
      t.Fatalf("tenc box content is too short: got %d bytes, want at least 8", len(tencBox.RemainingData))
   }
   perSampleIVSize = tencBox.RemainingData[7]
   if perSampleIVSize != 8 {
      t.Fatalf("Extracted incorrect perSampleIVSize: got %d, want 8", perSampleIVSize)
   }
   t.Logf("Successfully extracted perSampleIVSize = %d from init segment.", perSampleIVSize)

   // --- Step 2: Parse the Media Segment to find the raw 'senc' box ---
   var rawSenc *RawBox
   mediaParser := NewParser(sampleSegmentWithSenc)
   mediaBox, err := mediaParser.ParseNextBox()
   if err != nil || mediaBox.Moof == nil {
      t.Fatalf("Failed to parse moof box from media segment: %v", err)
   }

   var traf *TrafBox
   for _, child := range mediaBox.Moof.Children {
      if child.Traf != nil {
         traf = child.Traf
         break
      }
   }
   if traf == nil {
      t.Fatal("Failed to find 'traf' box in media segment")
   }

   for _, child := range traf.Children {
      if child.Raw != nil && child.Raw.Type == "senc" {
         rawSenc = child.Raw
         break
      }
   }
   if rawSenc == nil {
      t.Fatal("Failed to find raw 'senc' box in media segment")
   }
   t.Logf("Successfully extracted raw 'senc' box from media segment.")

   // --- Step 3: Call ParseSencContent with context and verify the results ---
   parsedSenc, err := ParseSencContent(rawSenc.Content, perSampleIVSize)
   if err != nil {
      t.Fatalf("ParseSencContent failed: %v", err)
   }
   if parsedSenc == nil {
      t.Fatal("ParseSencContent returned a nil box")
   }

   if parsedSenc.SampleCount != 2 {
      t.Errorf("Incorrect sample count: got %d, want 2", parsedSenc.SampleCount)
   }

   expectedIV1 := []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA}
   if !bytes.Equal(parsedSenc.InitializationVectors[0].IV, expectedIV1) {
      t.Errorf("Incorrect IV for sample 1: got %x, want %x", parsedSenc.InitializationVectors[0].IV, expectedIV1)
   }
   if len(parsedSenc.InitializationVectors[0].Subsamples) != 1 {
      t.Fatalf("Incorrect subsample count for sample 1: got %d, want 1", len(parsedSenc.InitializationVectors[0].Subsamples))
   }

   expectedIV2 := []byte{0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB}
   if !bytes.Equal(parsedSenc.InitializationVectors[1].IV, expectedIV2) {
      t.Errorf("Incorrect IV for sample 2: got %x, want %x", parsedSenc.InitializationVectors[1].IV, expectedIV2)
   }

   t.Log("Successfully parsed 'senc' content with context and verified data.")
}
