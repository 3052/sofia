# Overview

package `moov`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./moov.go#L12)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Box       []*sofia.Box
  Pssh      []pssh.Box
  Trak      trak.Box
}
```

ISO/IEC 14496-12
  aligned(8) class MovieBox extends Box('moov') {
  }

### func (\*Box) [Append](./moov.go#L19)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./moov.go#L39)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[moov.go](./moov.go)
