package sofia

import (
   "fmt"
   "log/slog"
   "os"
   "testing"
)

func TestOriginalFormat(t *testing.T) {
   h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
      Level: slog.LevelDebug,
   })
   slog.SetDefault(slog.New(h))
   media, err := os.Open("testdata/hulu-ec-3/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer media.Close()
   var f File
   if err := f.Read(media); err != nil {
      t.Fatal(err)
   }
   format := f.Movie.Track.Media.MediaInformation.SampleTable.SampleDescription.
      AudioSample.ProtectionScheme.OriginalFormat
   fmt.Printf("%q\n", format.DataFormat)
}
