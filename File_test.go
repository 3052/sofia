package sofia

import (
   "log/slog"
   "os"
   "testing"
)

func TestFile(t *testing.T) {
   in, err := os.Open("testdata/rtbf/vod-idx-video=4000000.dash")
   if err != nil {
      t.Fatal(err)
   }
   defer in.Close()
   var out File
   slog.SetLogLoggerLevel(slog.LevelDebug)
   err = out.Read(in)
   if err != nil {
      t.Fatal(err)
   }
}
