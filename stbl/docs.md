# Overview

package `stbl`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./stbl.go#L41)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Box       []sofia.Box
  Stsd      stsd.Box
}
```

ISO/IEC 14496-12
  aligned(8) class SampleTableBox extends Box('stbl') {
  }

### func (\*Box) [Append](./stbl.go#L47)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./stbl.go#L8)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[stbl.go](./stbl.go)
