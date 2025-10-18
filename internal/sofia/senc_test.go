package mp4parser

import (
   "bytes"
   "testing"
)

// TestParseSencWithContext simulates the full workflow of parsing an init segment for
// context and using that context to parse a media segment's 'senc' box.
func TestSenc(t *testing.T) {
   // --- Step 1: Parse the Initialization Segment to get context ---
   var perSampleIVSize uint8
   initParser := NewParser(sampleInitWithTenc)

   // In a real app, you would loop, but here we know the structure.
   moovBox, err := initParser.ParseNextBox()
   if err != nil || moovBox.Moov == nil {
      t.Fatalf("Failed to parse moov box from init segment: %v", err)
   }

   // This is a simplified navigation path; a real-world parser would loop and check children.
   trak := moovBox.Moov.Children[0].Trak
   stbl := trak.Children[0].Mdia.Children[0].Minf.Children[0].Stbl
   stsd := stbl.Children[0].Stsd
   encv := stsd.Children[0].Encv
   schi := encv.Children[0].Sinf.Children[0].Schi
   tencRaw := schi.Children[0].Raw // tenc is stored as a RawBox
   if tencRaw == nil || tencRaw.Type != "tenc" {
      t.Fatal("Failed to find 'tenc' box in the init segment")
   }

   // The default_per_sample_iv_size is at byte 7 (offset 6) of the tenc content payload.
   // tenc content: version(1), flags(3), reserved(1), is_protected(1), iv_size(1)
   if len(tencRaw.Content) < 7 {
      t.Fatalf("tenc box content is too short: got %d bytes, want at least 7", len(tencRaw.Content))
   }
   perSampleIVSize = tencRaw.Content[6]
   if perSampleIVSize != 8 {
      t.Fatalf("Extracted incorrect perSampleIVSize: got %d, want 8", perSampleIVSize)
   }
   t.Logf("Successfully extracted perSampleIVSize = %d from init segment.", perSampleIVSize)

   // --- Step 2: Parse the Media Segment to find the raw 'senc' box ---
   var rawSenc *RawBox
   mediaParser := NewParser(sampleSegmentWithSenc)
   moofBox, err := mediaParser.ParseNextBox()
   if err != nil || moofBox.Moof == nil {
      t.Fatalf("Failed to parse moof box from media segment: %v", err)
   }
   traf := moofBox.Moof.Children[0].Traf
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

   // Assertions to verify correct parsing
   if parsedSenc.SampleCount != 2 {
      t.Errorf("Incorrect sample count: got %d, want 2", parsedSenc.SampleCount)
   }
   if len(parsedSenc.InitializationVectors) != 2 {
      t.Errorf("Incorrect number of IVs parsed: got %d, want 2", len(parsedSenc.InitializationVectors))
   }

   // Check sample 1 data
   expectedIV1 := []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA}
   if !bytes.Equal(parsedSenc.InitializationVectors[0].IV, expectedIV1) {
      t.Errorf("Incorrect IV for sample 1: got %x, want %x", parsedSenc.InitializationVectors[0].IV, expectedIV1)
   }
   if len(parsedSenc.InitializationVectors[0].Subsamples) != 1 {
      t.Fatalf("Incorrect subsample count for sample 1: got %d, want 1", len(parsedSenc.InitializationVectors[0].Subsamples))
   }
   subsample1 := parsedSenc.InitializationVectors[0].Subsamples[0]
   if subsample1.BytesOfClearData != 0x1234 {
      t.Errorf("Incorrect clear data bytes for sample 1: got %d, want %d", subsample1.BytesOfClearData, 0x1234)
   }
   if subsample1.BytesOfProtectedData != 0x5678 {
      t.Errorf("Incorrect protected data bytes for sample 1: got %d, want %d", subsample1.BytesOfProtectedData, 0x5678)
   }

   // Check sample 2 data
   expectedIV2 := []byte{0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB}
   if !bytes.Equal(parsedSenc.InitializationVectors[1].IV, expectedIV2) {
      t.Errorf("Incorrect IV for sample 2: got %x, want %x", parsedSenc.InitializationVectors[1].IV, expectedIV2)
   }

   t.Log("Successfully parsed 'senc' content with context and verified data.")
}

// sampleInitWithTenc defines a minimal valid fMP4 initialization segment.
// It contains the critical path to the 'tenc' box to provide context.
var sampleInitWithTenc = []byte{
   // moov box
   0x00, 0x00, 0x00, 0x90, 'm', 'o', 'o', 'v',
   // trak box
   0x00, 0x00, 0x00, 0x88, 't', 'r', 'a', 'k',
   // mdia box
   0x00, 0x00, 0x00, 0x80, 'm', 'd', 'i', 'a',
   // minf box
   0x00, 0x00, 0x00, 0x78, 'm', 'i', 'n', 'f',
   // stbl box
   0x00, 0x00, 0x00, 0x70, 's', 't', 'b', 'l',
   // stsd box
   0x00, 0x00, 0x00, 0x68, 's', 't', 's', 'd',
   0x00, 0x00, 0x00, 0x00, // version/flags
   0x00, 0x00, 0x00, 0x01, // entry_count = 1
   // encv box (simplified, only contains sinf)
   0x00, 0x00, 0x00, 0x58, 'e', 'n', 'c', 'v',
   // ... 78 bytes of prefix data would go here ...
   // sinf box
   0x00, 0x00, 0x00, 0x50, 's', 'i', 'n', 'f',
   // schi box
   0x00, 0x00, 0x00, 0x28, 's', 'c', 'h', 'i',
   // tenc box (size 32)
   0x00, 0x00, 0x00, 0x20, 't', 'e', 'n', 'c',
   0x00, 0x00, 0x00, 0x00, // version/flags
   0x00, 0x01, // reserved, is_protected=1
   0x08,                                           // default_per_sample_iv_size = 8 (THE CRITICAL VALUE)
   0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, // default_KID
   0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
}

// sampleSegmentWithSenc defines a minimal media segment containing a raw 'senc' box.
var sampleSegmentWithSenc = []byte{
   // moof box
   0x00, 0x00, 0x00, 0x48, 'm', 'o', 'o', 'f',
   // traf box
   0x00, 0x00, 0x00, 0x40, 't', 'r', 'a', 'f',
   // senc box (size 32)
   0x00, 0x00, 0x00, 0x20, 's', 'e', 'n', 'c',
   0x00, 0x00, 0x00, 0x02, // version/flags (subsample_data_present)
   0x00, 0x00, 0x00, 0x02, // sample_count = 2
   // Sample 1
   0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, // IV (8 bytes because tenc said so)
   0x00, 0x01, // subsample_count = 1
   0x12, 0x34, // bytes_of_clear_data = 0x1234
   0x00, 0x00, 0x56, 0x78, // bytes_of_protected_data = 0x5678
   // Sample 2
   0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, // IV (8 bytes)
   // No subsamples for sample 2 in this test case
}
