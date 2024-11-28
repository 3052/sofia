# Overview

package `moof`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./moof.go#L12)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Box       []sofia.Box
  Traf      traf.Box
}
```

ISO/IEC 14496-12
  aligned(8) class MovieFragmentBox extends Box('moof') {
  }

### func (\*Box) [Append](./moof.go#L44)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./moof.go#L18)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[moof.go](./moof.go)
