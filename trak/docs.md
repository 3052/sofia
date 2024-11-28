# Overview

package `trak`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./trak.go#L11)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Box       []sofia.Box
  Mdia      mdia.Box
}
```

ISO/IEC 14496-12
  aligned(8) class TrackBox extends Box('trak') {
  }

### func (\*Box) [Append](./trak.go#L17)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./trak.go#L31)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[trak.go](./trak.go)
