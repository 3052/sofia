# Overview

package `traf`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./traf.go#L74)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Box       []*sofia.Box
  Senc      *senc.Box
  Tfhd      tfhd.Box
  Trun      trun.Box
}
```

ISO/IEC 14496-12
  aligned(8) class TrackFragmentBox extends Box('traf') {
  }

### func (\*Box) [Append](./traf.go#L82)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./traf.go#L11)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[traf.go](./traf.go)
