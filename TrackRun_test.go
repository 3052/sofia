package sofia

import (
   "log/slog"
   "os"
   "testing"
)

func TestTrackRun(t *testing.T) {
   h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
      Level: slog.LevelDebug,
   })
   slog.SetDefault(slog.New(h))
   seg, err := os.Open("testdata/paramount-avc1/seg_1.m4s")
   if err != nil {
      t.Fatal(err)
   }
   defer seg.Close()
   var value File
   err = value.Read(seg)
   if err != nil {
      t.Fatal(err)
   }
}
