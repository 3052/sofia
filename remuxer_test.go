package sofia

import (
   "bytes"
   "encoding/binary"
   "io"
   "os"
   "testing"
)

// TestRemuxer_Integration simulates a full cycle with synthetic data.
func TestRemuxer_Integration(t *testing.T) {
   outFile, err := os.CreateTemp("", "output_*.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer os.Remove(outFile.Name())
   defer outFile.Close()

   // Create Synthetic Data
   initSeg := createSyntheticInit()
   seg1 := createSyntheticSegment(1, []byte{0xAA, 0xAA, 0xAA, 0xAA})
   seg2 := createSyntheticSegment(2, []byte{0xBB, 0xBB, 0xBB, 0xBB})

   remuxer := &Remuxer{Writer: outFile}

   if err := remuxer.Initialize(initSeg); err != nil {
      t.Fatalf("Initialize failed: %v", err)
   }
   if err := remuxer.AddSegment(seg1); err != nil {
      t.Fatalf("AddSegment 1 failed: %v", err)
   }
   if err := remuxer.AddSegment(seg2); err != nil {
      t.Fatalf("AddSegment 2 failed: %v", err)
   }
   if err := remuxer.Finish(); err != nil {
      t.Fatalf("Finish failed: %v", err)
   }

   // Verify Output
   if _, err := outFile.Seek(0, io.SeekStart); err != nil {
      t.Fatal(err)
   }
   content, err := io.ReadAll(outFile)
   if err != nil {
      t.Fatal(err)
   }

   // Check mdat payload
   expectedPayload := []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB}
   if !bytes.Contains(content, expectedPayload) {
      t.Error("Output missing concatenated mdat payloads")
   }

   // Check moov presence
   if !bytes.Contains(content, []byte("moov")) {
      t.Error("Output missing 'moov' box")
   }

   t.Logf("Successfully created MP4 of size %d bytes", len(content))
}

// --- Test Helpers ---

func createSyntheticInit() []byte {
   ftyp := makeBox("ftyp", []byte("iso50000"))

   mvhdData := make([]byte, 108)
   binary.BigEndian.PutUint32(mvhdData[20:24], 1000)
   mvhd := makeBox("mvhd", mvhdData)

   mdhdData := make([]byte, 32)
   binary.BigEndian.PutUint32(mdhdData[20:24], 1000)
   mdhd := makeBox("mdhd", mdhdData)

   stsd := makeBox("stsd", make([]byte, 8))
   stbl := makeBox("stbl", stsd)
   minf := makeBox("minf", stbl)
   mdia := makeBox("mdia", append(mdhd, minf...))
   trak := makeBox("trak", mdia)

   moov := makeBox("moov", append(mvhd, trak...))

   return append(ftyp, moov...)
}

func createSyntheticSegment(seq int, payload []byte) []byte {
   tfhdData := make([]byte, 8)
   binary.BigEndian.PutUint32(tfhdData[4:8], 1)
   tfhd := makeBox("tfhd", tfhdData)

   trunFlags := uint32(0x000301)
   sampleCount := uint32(1)

   trunData := make([]byte, 8)
   binary.BigEndian.PutUint32(trunData[0:4], trunFlags)
   binary.BigEndian.PutUint32(trunData[4:8], sampleCount)
   trunData = append(trunData, 0, 0, 0, 0) // data offset
   trunData = append(trunData, uint32ToBytes(100)...)
   trunData = append(trunData, uint32ToBytes(uint32(len(payload)))...)
   trun := makeBox("trun", trunData)

   traf := makeBox("traf", append(tfhd, trun...))
   moof := makeBox("moof", traf)
   mdat := makeBox("mdat", payload)

   return append(moof, mdat...)
}

func makeBox(typeStr string, payload []byte) []byte {
   size := 8 + len(payload)
   buffer := make([]byte, 8)
   binary.BigEndian.PutUint32(buffer[0:4], uint32(size))
   copy(buffer[4:8], []byte(typeStr))
   return append(buffer, payload...)
}

func uint32ToBytes(v uint32) []byte {
   buffer := make([]byte, 4)
   binary.BigEndian.PutUint32(buffer, v)
   return buffer
}
