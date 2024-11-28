# Overview

package `schi`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./schi.go#L12)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Tenc      tenc.Box
}
```

ISO/IEC 14496-12
  aligned(8) class SchemeInformationBox extends Box('schi') {
     Box scheme_specific_data[];
  }

### func (\*Box) [Append](./schi.go#L17)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./schi.go#L25)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[schi.go](./schi.go)
