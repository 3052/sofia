# Overview

package `mdia`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./mdia.go#L11)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Box       []sofia.Box
  Minf      minf.Box
}
```

ISO/IEC 14496-12
  aligned(8) class MediaBox extends Box('mdia') {
  }

### func (\*Box) [Append](./mdia.go#L17)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./mdia.go#L31)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[mdia.go](./mdia.go)
