package mp4parser

import (
   "bytes"
   "os"
   "testing"
)

func TestRoundtrip(t *testing.T) {
   sampleFMP4, err := os.ReadFile("pts_0.mp4")
   if err != nil {
      t.Fatal(err)
   }
   parser := NewParser(sampleFMP4)
   var parsedBoxes []*Box
   var totalParsedSize uint64
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
      t.Errorf("Parser did not consume the entire file: got %d bytes, want %d bytes", totalParsedSize, len(sampleFMP4))
   }
   formattedBuffer := new(bytes.Buffer)
   for _, box := range parsedBoxes {
      formattedBytes, err := box.Format()
      if err != nil {
         t.Fatalf("Failed to format box of type '%s': %v", box.Header.Type, err)
      }
      formattedBuffer.Write(formattedBytes)
   }
   formattedData := formattedBuffer.Bytes()
   if len(sampleFMP4) != len(formattedData) {
      t.Fatalf("Length mismatch: original is %d bytes, formatted is %d bytes", len(sampleFMP4), len(formattedData))
   }
   if !bytes.Equal(sampleFMP4, formattedData) {
      t.Errorf("Roundtrip failed: formatted data does not match original data")
   }
   t.Logf("Successfully roundtripped %d bytes.", len(sampleFMP4))
}
