package sofia

import (
   "bytes"
   "crypto/aes"
   "crypto/cipher"
   "io"
   "os"
   "testing"
)

// TestUnfragmenter_Decryption verifies the OnSample callback can decrypt data.
func TestUnfragmenter_Decryption(t *testing.T) {
   key := []byte("0123456789abcdef")
   iv := []byte("12345678")
   plaintext := []byte("SecretPayload!")
   encrypted := make([]byte, len(plaintext))

   // Encrypt manually
   block, _ := aes.NewCipher(key)
   iv16 := make([]byte, 16)
   copy(iv16, iv)
   stream := cipher.NewCTR(block, iv16)
   stream.XORKeyStream(encrypted, plaintext)

   outFile, err := os.CreateTemp("", "dec_out_*.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer os.Remove(outFile.Name())
   defer outFile.Close()

   unfrag := &Unfragmenter{
      Writer: outFile,
      OnSample: func(sample []byte, encInfo *SampleEncryptionInfo) {
         if encInfo != nil {
            DecryptSample(sample, encInfo, block)
         }
      },
   }

   initSeg := createSyntheticInit()
   if err := unfrag.Initialize(initSeg); err != nil {
      t.Fatal(err)
   }

   // Create a segment with a 'senc' box and encrypted mdat
   seg := createSyntheticEncryptedSegment(encrypted, iv)
   if err := unfrag.AddSegment(seg); err != nil {
      t.Fatal(err)
   }

   if err := unfrag.Finish(); err != nil {
      t.Fatal(err)
   }

   // Read back and check if payload matches plaintext
   if _, err := outFile.Seek(0, io.SeekStart); err != nil {
      t.Fatal(err)
   }
   data, err := io.ReadAll(outFile)
   if err != nil {
      t.Fatal(err)
   }

   if !bytes.Contains(data, plaintext) {
      t.Error("Output does not contain decrypted plaintext")
   }
}
