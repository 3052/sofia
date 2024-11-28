# Overview

package `sidx`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) GetSize() int](#func-box-getsize)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
  - [type Reference](#type-reference)
    - [func (r Reference) Append(data []byte) ([]byte, error)](#func-reference-append)
    - [func (r \*Reference) Decode(data []byte) (int, error)](#func-reference-decode)
    - [func (r \*Reference) SetSize(size uint32)](#func-reference-setsize)
    - [func (r Reference) Size() uint32](#func-reference-size)
- [Source files](#source-files)

## Types

### type [Box](./sidx.go#L66)

```go
type Box struct {
  BoxHeader                sofia.BoxHeader
  FullBoxHeader            sofia.FullBoxHeader
  ReferenceId              uint32
  Timescale                uint32
  EarliestPresentationTime []byte
  FirstOffset              []byte
  Reserved                 uint16
  ReferenceCount           uint16
  Reference                []Reference
}
```

ISO/IEC 14496-12
  aligned(8) class SegmentIndexBox extends FullBox('sidx', version, 0) {
     unsigned int(32) reference_ID;
     unsigned int(32) timescale;
     if (version==0) {
        unsigned int(32) earliest_presentation_time;
        unsigned int(32) first_offset;
     } else {
        unsigned int(64) earliest_presentation_time;
        unsigned int(64) first_offset;
     }
     unsigned int(16) reserved = 0;
     unsigned int(16) reference_count;
     for(i=1; i <= reference_count; i++) {
        bit (1) reference_type;
        unsigned int(31) referenced_size;
        unsigned int(32) subsegment_duration;
        bit(1) starts_with_SAP;
        unsigned int(3) SAP_type;
        unsigned int(28) SAP_delta_time;
     }
  }

### func (\*Box) [Append](./sidx.go#L20)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [GetSize](./sidx.go#L8)

```go
func (b *Box) GetSize() int
```

### func (\*Box) [Read](./sidx.go#L78)

```go
func (b *Box) Read(data []byte) error
```

### type [Reference](./sidx.go#L130)

```go
type Reference [3]uint32
```

### func (Reference) [Append](./sidx.go#L132)

```go
func (r Reference) Append(data []byte) ([]byte, error)
```

### func (\*Reference) [Decode](./sidx.go#L136)

```go
func (r *Reference) Decode(data []byte) (int, error)
```

### func (\*Reference) [SetSize](./sidx.go#L125)

```go
func (r *Reference) SetSize(size uint32)
```

### func (Reference) [Size](./sidx.go#L145)

```go
func (r Reference) Size() uint32
```

this is the size of the fragment, typically `moof` + `mdat`

## Source files

[sidx.go](./sidx.go)
