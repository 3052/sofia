# Overview

package `enca`

## Index

- [Types](#types)
  - [type SampleEntry](#type-sampleentry)
    - [func (s \*SampleEntry) Append(data []byte) ([]byte, error)](#func-sampleentry-append)
    - [func (s \*SampleEntry) Read(data []byte) error](#func-sampleentry-read)
- [Source files](#source-files)

## Types

### type [SampleEntry](./enca.go#L18)

```go
type SampleEntry struct {
  SampleEntry sofia.SampleEntry
  Extends     struct {
    ChannelCount uint16
    SampleSize   uint16
    PreDefined   uint16

    SampleRate uint32
    // contains filtered or unexported fields
  }
  Box  []*sofia.Box
  Sinf sinf.Box
}
```

ISO/IEC 14496-12
  class AudioSampleEntry(codingname) extends SampleEntry(codingname) {
     const unsigned int(32)[2] reserved = 0;
     unsigned int(16) channelcount;
     template unsigned int(16) samplesize = 16;
     unsigned int(16) pre_defined = 0;
     const unsigned int(16) reserved = 0 ;
     template unsigned int(32) samplerate = { default samplerate of media}<<16;
  }

### func (\*SampleEntry) [Append](./enca.go#L32)

```go
func (s *SampleEntry) Append(data []byte) ([]byte, error)
```

### func (\*SampleEntry) [Read](./enca.go#L50)

```go
func (s *SampleEntry) Read(data []byte) error
```

## Source files

[enca.go](./enca.go)
