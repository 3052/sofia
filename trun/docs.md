# Overview

package `trun`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
  - [type Sample](#type-sample)
    - [func (s \*Sample) Append(data []byte) ([]byte, error)](#func-sample-append)
    - [func (s \*Sample) Decode(data []byte) (int, error)](#func-sample-decode)
- [Source files](#source-files)

## Types

### type [Box](./trun.go#L31)

```go
type Box struct {
  BoxHeader        sofia.BoxHeader
  FullBoxHeader    sofia.FullBoxHeader
  SampleCount      uint32
  DataOffset       int32
  FirstSampleFlags uint32
  Sample           []Sample
}
```

ISO/IEC 14496-12

If the data-offset is present, it is relative to the base-data-offset
established in the track fragment header.

sample-size-present: each sample has its own size, otherwise the default is
used.

  aligned(8) class TrackRunBox extends FullBox('trun', version, tr_flags) {
     unsigned int(32) sample_count;
     signed int(32) data_offset; // 0x000001, assume present
     unsigned int(32) first_sample_flags; // 0x000004
     {
        unsigned int(32) sample_duration; // 0x000100
        unsigned int(32) sample_size; // 0x000200
        unsigned int(32) sample_flags // 0x000400
        if (version == 0) {
           unsigned int(32) sample_composition_time_offset; // 0x000800
        } else {
           signed int(32) sample_composition_time_offset; // 0x000800
        }
     }[ sample_count ]
  }

### func (\*Box) [Append](./trun.go#L40)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./trun.go#L91)

```go
func (b *Box) Read(data []byte) error
```

### type [Sample](./trun.go#L172)

```go
type Sample struct {
  Duration              uint32
  SampleSize            uint32
  Flags                 uint32
  CompositionTimeOffset [4]byte
  // contains filtered or unexported fields
}
```

### func (\*Sample) [Append](./trun.go#L156)

```go
func (s *Sample) Append(data []byte) ([]byte, error)
```

### func (\*Sample) [Decode](./trun.go#L127)

```go
func (s *Sample) Decode(data []byte) (int, error)
```

## Source files

[trun.go](./trun.go)
