package sofia

import (
   "bytes"
   "crypto/aes"
   "encoding/binary"
   "encoding/hex"
   "io"
   "os"
   "path/filepath"
   "sort"
   "testing"
)

// TestUnfragmenter_Integration simulates a full cycle with synthetic data.
func TestUnfragmenter_Integration(t *testing.T) {
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

   unfrag := &Unfragmenter{Writer: outFile}

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

// TestUnfragmenter_RealFiles looks for real files in the "ignore" directory.
func TestUnfragmenter_RealFiles(t *testing.T) {
   workDir := "ignore"
   if _, err := os.Stat(workDir); os.IsNotExist(err) {
      t.Fatalf("directory '%s' not found", workDir)
   }

   outPath := filepath.Join(workDir, "joined_output.mp4")
   outFile, err := os.Create(outPath)
   if err != nil {
      t.Fatalf("Failed to create output file: %v", err)
   }
   defer outFile.Close()

   key, err := hex.DecodeString("c35acf72e42c8a9ca31da21007a17a65")
   if err != nil {
      t.Fatalf("Failed to decode key: %v", err)
   }
   block, err := aes.NewCipher(key)
   if err != nil {
      t.Fatalf("Failed to create cipher: %v", err)
   }

   unfrag := &Unfragmenter{
      Writer: outFile,
      OnSample: func(sample []byte, encInfo *SampleEncryptionInfo) {
         DecryptSample(sample, encInfo, block)
      },
   }

   initPath := filepath.Join(workDir, "_init.mp4")
   initData, err := os.ReadFile(initPath)
   if err != nil {
      t.Fatalf("Failed to read init segment (%s): %v", initPath, err)
   }

   if err := unfrag.Initialize(initData); err != nil {
      t.Fatalf("Unfragmenter.Initialize failed: %v", err)
   }

   globPattern := filepath.Join(workDir, "*.m4s")
   segmentFiles, err := filepath.Glob(globPattern)
   if err != nil {
      t.Fatalf("Failed to glob segments: %v", err)
   }
   if len(segmentFiles) == 0 {
      t.Fatalf("No segment files found matching pattern: %s", globPattern)
   }
   sort.Strings(segmentFiles)

   t.Logf("Found %d segments. Processing...", len(segmentFiles))
   for _, segmentPath := range segmentFiles {
      segData, err := os.ReadFile(segmentPath)
      if err != nil {
         t.Fatalf("Failed to read segment %s: %v", segmentPath, err)
      }
      if err := unfrag.AddSegment(segData); err != nil {
         t.Fatalf("Failed to add segment %s: %v", segmentPath, err)
      }
   }

   if err := unfrag.Finish(); err != nil {
      t.Fatalf("Unfragmenter.Finish failed: %v", err)
   }

   stat, _ := outFile.Stat()
   t.Logf("Success! Created %s (%d bytes)", outPath, stat.Size())
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
   buf := make([]byte, 8)
   binary.BigEndian.PutUint32(buf[0:4], uint32(size))
   copy(buf[4:8], []byte(typeStr))
   return append(buf, payload...)
}

func uint32ToBytes(v uint32) []byte {
   b := make([]byte, 4)
   binary.BigEndian.PutUint32(b, v)
   return b
}
