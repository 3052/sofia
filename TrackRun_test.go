package sofia

import (
   "log/slog"
   "os"
   "testing"
)

func Test_Trun(t *testing.T) {
   h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
      Level: slog.LevelDebug,
   })
   slog.SetDefault(slog.New(h))
   seg, err := os.Open("testdata/paramount-video/seg_1303.m4s")
   if err != nil {
      t.Fatal(err)
   }
   defer seg.Close()
   var f File
   if err := f.Decode(seg); err != nil {
      t.Fatal(err)
   }
}
