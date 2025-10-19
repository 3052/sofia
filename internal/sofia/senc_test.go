package mp4parser

import (
   "bytes"
   "testing"
)

// buildTestInitSegment programmatically creates a valid init segment byte slice.
func buildTestInitSegment() []byte {
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
   trakBox := &TrakBox{Children: []*TrakChildBox{{Mdia: mdiaBox}}}
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
   0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02,
   0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
   0x00, 0x01, 0x12, 0x34, 0x00, 0x00, 0x56, 0x78,
   0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
   0x00, 0x00,
}

// --- Navigation Helper Functions ---

func findTrak(moov *MoovBox) *TrakBox {
   for _, child := range moov.Children {
      if child.Trak != nil {
         return child.Trak
      }
   }
   return nil
}

func findMdia(trak *TrakBox) *MdiaBox {
   for _, child := range trak.Children {
      if child.Mdia != nil {
         return child.Mdia
      }
   }
   return nil
}

func findMinf(mdia *MdiaBox) *MinfBox {
   for _, child := range mdia.Children {
      if child.Minf != nil {
         return child.Minf
      }
   }
   return nil
}

func findStbl(minf *MinfBox) *StblBox {
   for _, child := range minf.Children {
      if child.Stbl != nil {
         return child.Stbl
      }
   }
   return nil
}

func findStsd(stbl *StblBox) *StsdBox {
   for _, child := range stbl.Children {
      if child.Stsd != nil {
         return child.Stsd
      }
   }
   return nil
}

func findEncv(stsd *StsdBox) *EncvBox {
   for _, child := range stsd.Children {
      if child.Encv != nil {
         return child.Encv
      }
   }
   return nil
}

func findSinf(encv *EncvBox) *SinfBox {
   for _, child := range encv.Children {
      if child.Sinf != nil {
         return child.Sinf
      }
   }
   return nil
}

func findSchi(sinf *SinfBox) *SchiBox {
   for _, child := range sinf.Children {
      if child.Schi != nil {
         return child.Schi
      }
   }
   return nil
}

func findTenc(schi *SchiBox) *TencBox {
   for _, child := range schi.Children {
      if child.Tenc != nil {
         return child.Tenc
      }
   }
   return nil
}

func TestParseSencWithContext(t *testing.T) {
   // --- Step 1: Parse the Initialization Segment and navigate to 'tenc' ---
   initParser := NewParser(sampleInitWithTenc)
   initBox, err := initParser.ParseNextBox()
   if err != nil || initBox.Moov == nil {
      t.Fatalf("Failed to parse moov box from init segment: %v", err)
   }

   // Use the helper functions to navigate cleanly, checking for errors at each step.
   trak := findTrak(initBox.Moov)
   if trak == nil {
      t.Fatal("Failed to find 'trak' box")
   }
   mdia := findMdia(trak)
   if mdia == nil {
      t.Fatal("Failed to find 'mdia' box")
   }
   minf := findMinf(mdia)
   if minf == nil {
      t.Fatal("Failed to find 'minf' box")
   }
   stbl := findStbl(minf)
   if stbl == nil {
      t.Fatal("Failed to find 'stbl' box")
   }
   stsd := findStsd(stbl)
   if stsd == nil {
      t.Fatal("Failed to find 'stsd' box")
   }
   encv := findEncv(stsd)
   if encv == nil {
      t.Fatal("Failed to find 'encv' box")
   }
   sinf := findSinf(encv)
   if sinf == nil {
      t.Fatal("Failed to find 'sinf' box")
   }
   schi := findSchi(sinf)
   if schi == nil {
      t.Fatal("Failed to find 'schi' box")
   }
   tencBox := findTenc(schi)
   if tencBox == nil {
      t.Fatal("Failed to find 'tenc' box in the init segment")
   }

   if len(tencBox.RemainingData) < 8 {
      t.Fatalf("tenc box content is too short: got %d bytes, want at least 8", len(tencBox.RemainingData))
   }
   perSampleIVSize := tencBox.RemainingData[7]
   if perSampleIVSize != 8 {
      t.Fatalf("Extracted incorrect perSampleIVSize: got %d, want 8", perSampleIVSize)
   }
   t.Logf("Successfully extracted perSampleIVSize = %d from init segment.", perSampleIVSize)

   // --- Step 2: Parse the Media Segment to find the raw 'senc' box ---
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

   var rawSenc *RawBox
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

   // Assertions
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
