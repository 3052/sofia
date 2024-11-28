# Overview

package `encv`

## Index

- [Types](#types)
  - [type SampleEntry](#type-sampleentry)
    - [func (s \*SampleEntry) Append(data []byte) ([]byte, error)](#func-sampleentry-append)
    - [func (s \*SampleEntry) Read(data []byte) error](#func-sampleentry-read)
- [Source files](#source-files)

## Types

### type [SampleEntry](./encv.go#L27)

```go
type SampleEntry struct {
  SampleEntry sofia.SampleEntry
  Extends     struct {
    Width           uint16
    Height          uint16
    HorizResolution uint32
    VertResolution  uint32

    FrameCount     uint16
    CompressorName [32]uint8
    Depth          uint16
    // contains filtered or unexported fields
  }
  Box  []*sofia.Box
  Sinf sinf.Box
}
```

ISO/IEC 14496-12
  class VisualSampleEntry(codingname) extends SampleEntry(codingname) {
     unsigned int(16) pre_defined = 0;
     const unsigned int(16) reserved = 0;
     unsigned int(32)[3] pre_defined = 0;
     unsigned int(16) width;
     unsigned int(16) height;
     template unsigned int(32) horizresolution = 0x00480000; // 72 dpi
     template unsigned int(32) vertresolution = 0x00480000; // 72 dpi
     const unsigned int(32) reserved = 0;
     template unsigned int(16) frame_count = 1;
     uint(8)[32] compressorname;
     template unsigned int(16) depth = 0x0018;
     int(16) pre_defined = -1;
     // other boxes from derived specifications
     CleanApertureBox clap; // optional
     PixelAspectRatioBox pasp; // optional
  }

### func (\*SampleEntry) [Append](./encv.go#L47)

```go
func (s *SampleEntry) Append(data []byte) ([]byte, error)
```

### func (\*SampleEntry) [Read](./encv.go#L65)

```go
func (s *SampleEntry) Read(data []byte) error
```

## Source files

[encv.go](./encv.go)
