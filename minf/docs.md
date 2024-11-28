# Overview

package `minf`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./minf.go#L11)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Box       []sofia.Box
  Stbl      stbl.Box
}
```

ISO/IEC 14496-12
  aligned(8) class MediaInformationBox extends Box('minf') {
  }

### func (\*Box) [Append](./minf.go#L17)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./minf.go#L31)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[minf.go](./minf.go)
