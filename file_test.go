package sofia

import (
	"os"
	"testing"
)

func Test_File(t *testing.T) {
	src, err := os.Open("testdata/amc-audio/segment0.m4f")
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst, err := os.Create("segment0.m4f")
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	var f File
	if err := f.Decode(src); err != nil {
		t.Fatal(err)
	}
	if err := f.Encode(dst); err != nil {
		t.Fatal(err)
	}
}
