package container

import (
   "log/slog"
   "os"
   "testing"
)

const file_test = "_init.mp4"

func TestFile(t *testing.T) {
   data, err := os.ReadFile(file_test)
   if err != nil {
      t.Fatal(err)
   }
   var video_eng File
   slog.SetLogLoggerLevel(slog.LevelDebug)
   err = video_eng.Read(data)
   if err != nil {
      t.Fatal(err)
   }
}
