package sofia

import (
	"encoding/json"
	"os"
	"testing"
)

func Test_Senc(t *testing.T) {
	media, err := os.Open("testdata/roku-video/index_video_8_0_1.mp4")
	if err != nil {
		t.Fatal(err)
	}
	defer media.Close()
	var f File
	if err := f.Decode(media); err != nil {
		t.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(f.Moof.Traf.Senc)
}
