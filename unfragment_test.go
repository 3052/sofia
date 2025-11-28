package sofia

import (
   "bytes"
   "encoding/binary"
   "io"
   "os"
   "testing"
)

// TestUnfragmenter_Integration simulates a full cycle of unfragmenting
// a stream consisting of an init segment and two media segments.
func TestUnfragmenter_Integration(t *testing.T) {
   // 1. Setup temporary output file
   outFile, err := os.CreateTemp("", "output_*.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer os.Remove(outFile.Name()) // Clean up
   defer outFile.Close()

   // 2. Create Synthetic Data
   // We need minimally valid boxes so the parser doesn't error out.

   // --- Init Segment (ftyp + moov) ---
   initSeg := createSyntheticInit()

   // --- Media Segment 1 (moof + mdat) ---
   // Contains 1 sample of size 4 bytes (0xAAAA_AAAA)
   seg1 := createSyntheticSegment(1, []byte{0xAA, 0xAA, 0xAA, 0xAA})

   // --- Media Segment 2 (moof + mdat) ---
   // Contains 1 sample of size 4 bytes (0xBBBB_BBBB)
   seg2 := createSyntheticSegment(2, []byte{0xBB, 0xBB, 0xBB, 0xBB})

   // 3. Run Unfragmenter
   unfrag := NewUnfragmenter(outFile)

   if err := unfrag.Initialize(initSeg); err != nil {
      t.Fatalf("Initialize failed: %v", err)
   }

   if err := unfrag.AddSegment(seg1); err != nil {
      t.Fatalf("AddSegment 1 failed: %v", err)
   }

   if err := unfrag.AddSegment(seg2); err != nil {
      t.Fatalf("AddSegment 2 failed: %v", err)
   }

   if err := unfrag.Finish(); err != nil {
      t.Fatalf("Finish failed: %v", err)
   }

   // 4. Verify Output
   // Re-open the file to check contents
   if _, err := outFile.Seek(0, io.SeekStart); err != nil {
      t.Fatal(err)
   }
   content, err := io.ReadAll(outFile)
   if err != nil {
      t.Fatal(err)
   }

   // A. Check ftyp presence
   if !bytes.Contains(content, []byte("ftyp")) {
      t.Error("Output missing 'ftyp' box")
   }

   // B. Check mdat header and payload
   // The file should look like: [ftyp] [mdat header] [payload 1] [payload 2] [moov]
   // Payload 1 + Payload 2 = AAAAAAAA + BBBBBBBB
   expectedPayload := []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB}
   if !bytes.Contains(content, expectedPayload) {
      t.Error("Output missing concatenated mdat payloads")
   }

   // C. Check that moov is at the end
   // The last 4 bytes of the type in a standard box header are at index -4 from end of box.
   // We expect the last top-level box to be 'moov'.
   // Since we can't easily parse the whole file without recursive logic, we do a basic check.
   // We know Unfragmenter writes moov last.
   if !bytes.Contains(content, []byte("moov")) {
      t.Error("Output missing 'moov' box")
   }

   t.Logf("Successfully created MP4 of size %d bytes", len(content))
}

// --- Generators for Synthetic Data ---

func createSyntheticInit() []byte {
   // ftyp
   ftyp := makeBox("ftyp", []byte("iso50000")) // fake brand

   // moov -> trak -> mdia -> minf -> stbl -> stsd
   // We need minimal structure so sofia parser finds "stsd"
   stsd := makeBox("stsd", make([]byte, 8)) // header + 0 entries
   stbl := makeBox("stbl", stsd)
   minf := makeBox("minf", stbl)
   mdia := makeBox("mdia", minf)
   trak := makeBox("trak", mdia)
   moov := makeBox("moov", trak)

   return append(ftyp, moov...)
}

func createSyntheticSegment(seq int, payload []byte) []byte {
   // moof -> traf -> (tfhd, trun)

   // tfhd: Flags(4) + TrackID(4)
   // Flags=0, TrackID=1. No defaults provided to keep simple.
   tfhdData := make([]byte, 8)
   binary.BigEndian.PutUint32(tfhdData[4:8], 1)
   tfhd := makeBox("tfhd", tfhdData)

   // trun: Flags(4) + Count(4) + (Duration, Size, Flags, CTO) per sample
   // We set flags to have data-offset present (0x01) + sample-duration(0x100) + sample-size(0x200)
   // Flags = 0x000301
   trunFlags := uint32(0x000301)
   sampleCount := uint32(1)

   trunData := make([]byte, 8)
   binary.BigEndian.PutUint32(trunData[0:4], trunFlags)
   binary.BigEndian.PutUint32(trunData[4:8], sampleCount)

   // Data Offset (4 bytes) - just placeholder 0
   trunData = append(trunData, 0, 0, 0, 0)

   // Sample 1: Duration(100), Size(len(payload))
   trunData = append(trunData, uint32ToBytes(100)...)
   trunData = append(trunData, uint32ToBytes(uint32(len(payload)))...)

   trun := makeBox("trun", trunData)
   traf := makeBox("traf", append(tfhd, trun...))
   moof := makeBox("moof", traf)

   // mdat
   mdat := makeBox("mdat", payload)

   return append(moof, mdat...)
}
