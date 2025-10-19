// File: senc_test.go
package mp4parser

import (
	"bytes"
	"encoding/hex"
	"log"
	"os"
	"testing"
)

// ... (the rest of the test file, including findTencBox and findSencContent, remains the same)
func TestParseSencFromFiles(t *testing.T) {
	// --- Step 1: Parse the initialization file to find the 'tenc' box ---
	initData, err := os.ReadFile("video_eng=108536.dash")
	if err != nil {
		t.Fatalf("Failed to read initialization file 'video_eng=108536.dash': %v."+
			"\nPlease ensure the file is in the same directory as the test.", err)
	}

	var tencBox *TencBox
	initParser := NewParser(initData)

	for initParser.HasMore() {
		box, err := initParser.ParseNextBox()
		if err != nil {
			t.Fatalf("Error parsing initialization file: %v", err)
		}
		if box != nil && box.Moov != nil {
			tencBox = findTencBox(box.Moov)
			break
		}
	}

	if tencBox == nil {
		t.Fatalf("Failed to find 'tenc' box in the initialization file.")
	}

	// --- Step 2: Parse the media segment file to find the 'senc' box content ---
	mediaData, err := os.ReadFile("video_eng=108536-0.dash")
	if err != nil {
		t.Fatalf("Failed to read media segment file 'video_eng=108536-0.dash': %v."+
			"\nPlease ensure the file is in the same directory as the test.", err)
	}

	var sencContent []byte
	mediaParser := NewParser(mediaData)

	for mediaParser.HasMore() {
		mediaBox, err := mediaParser.ParseNextBox()
		if err != nil {
			t.Fatalf("Error parsing media segment: %v", err)
		}
		if mediaBox == nil {
			break
		}
		if mediaBox.Moof != nil {
			sencContent = findSencContent(mediaBox.Moof)
			if sencContent != nil {
				break
			}
		}
	}

	if sencContent == nil {
		t.Fatalf("Failed to find 'senc' box content in the media segment file.")
	}

	// LOGGING: Confirm the size of the slice before passing it to the parser
	log.Printf("[TestParseSencFromFiles] Found 'senc' content with length: %d. Now attempting to parse...", len(sencContent))

	// --- Step 3: Call ParseSencContent with the extracted data ---
	sencBox, err := ParseSencContent(sencContent, tencBox.DefaultPerSampleIVSize, tencBox.DefaultConstantIV)
	if err != nil {
		t.Fatalf("ParseSencContent failed: %v", err)
	}

	// --- Step 4: Verify the results based on the provided diagrams ---
	expectedSampleCount := uint32(50)
	if sencBox.SampleCount != expectedSampleCount {
		t.Errorf("Expected SampleCount to be %d, but got %d", expectedSampleCount, sencBox.SampleCount)
	}

	if len(sencBox.InitializationVectors) != int(expectedSampleCount) {
		t.Fatalf("Expected %d InitializationVectors, but got %d", expectedSampleCount, len(sencBox.InitializationVectors))
	}

	expectedConstantIV, _ := hex.DecodeString("fbef035cb3b54819a1a3c213aeff15b2")
	if !bytes.Equal(tencBox.DefaultConstantIV, expectedConstantIV) {
		t.Errorf("Parsed incorrect DefaultConstantIV. Expected %x, got %x",
			expectedConstantIV, tencBox.DefaultConstantIV)
	}

	for i, iv := range sencBox.InitializationVectors {
		if !bytes.Equal(iv.IV, tencBox.DefaultConstantIV) {
			t.Errorf("Sample %d: Expected IV to be %x, but got %x",
				i, tencBox.DefaultConstantIV, iv.IV)
		}
	}

	t.Logf("Successfully parsed 'senc' box for %d samples.", sencBox.SampleCount)
	t.Logf("All samples correctly use the DefaultConstantIV from the 'tenc' box.")
}

func findTencBox(moov *MoovBox) *TencBox {
	for _, moovChild := range moov.Children {
		if moovChild.Trak != nil {
			for _, trakChild := range moovChild.Trak.Children {
				if trakChild.Mdia != nil {
					for _, mdiaChild := range trakChild.Mdia.Children {
						if mdiaChild.Minf != nil {
							for _, minfChild := range mdiaChild.Minf.Children {
								if minfChild.Stbl != nil {
									for _, stblChild := range minfChild.Stbl.Children {
										if stblChild.Stsd != nil {
											for _, stsdChild := range stblChild.Stsd.Children {
												if stsdChild.Encv != nil {
													for _, encvChild := range stsdChild.Encv.Children {
														if encvChild.Sinf != nil {
															for _, sinfChild := range encvChild.Sinf.Children {
																if sinfChild.Schi != nil {
																	for _, schiChild := range sinfChild.Schi.Children {
																		if schiChild.Tenc != nil {
																			return schiChild.Tenc
																		}
																	}
																}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func findSencContent(moof *MoofBox) []byte {
	for _, moofChild := range moof.Children {
		if moofChild.Traf != nil {
			for _, trafChild := range moofChild.Traf.Children {
				if trafChild.Senc != nil {
					return trafChild.Senc.Content
				}
			}
		}
	}
	return nil
}
