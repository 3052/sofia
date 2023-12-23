package sofia

import (
	"fmt"
	"os"
	"testing"
)

func Test_Trun(t *testing.T) {
	media, err := os.Open("testdata/amc-video/segment0.m4f")
	if err != nil {
		t.Fatal(err)
	}
	defer media.Close()
	var f File
	if err := f.Decode(media); err != nil {
		t.Fatal(err)
	}
	for _, sample := range f.Moof.Traf.Trun.Samples {
		fmt.Println(sample.Size)
	}
	fmt.Println("len(f.Moof.Traf.Trun.Samples)", len(f.Moof.Traf.Trun.Samples))
}
