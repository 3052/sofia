package sofia

import (
   "log/slog"
   "os"
   "testing"
)

func TestFile(t *testing.T) {
   //testdata\max-dvh1\segment-1.0001.m4s
   in, err := os.Open("testdata/max-dvh1/init.mp4")
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
