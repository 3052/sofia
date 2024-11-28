# Overview

package `tfhd`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./tfhd.go#L95)

```go
type Box struct {
  BoxHeader              sofia.BoxHeader
  FullBoxHeader          sofia.FullBoxHeader
  TrackId                uint32
  BaseDataOffset         uint64
  SampleDescriptionIndex uint32
  DefaultSampleDuration  uint32
  DefaultSampleSize      uint32
  DefaultSampleFlags     uint32
}
```

ISO/IEC 14496-12

  aligned(8) class TrackFragmentHeaderBox extends FullBox(
     'tfhd', 0, tf_flags
  ) {
     unsigned int(32) track_ID;
     // all the following are optional fields
     // their presence is indicated by bits in the tf_flags
     unsigned int(64) base_data_offset;
     unsigned int(32) sample_description_index;
     unsigned int(32) default_sample_duration;
     unsigned int(32) default_sample_size;
     unsigned int(32) default_sample_flags;
  }

### func (\*Box) [Append](./tfhd.go#L106)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./tfhd.go#L33)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[tfhd.go](./tfhd.go)
