# Overview

package `sinf`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./sinf.go#L15)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Box       []sofia.Box
  Frma      frma.Box
  Schi      schi.Box
}
```

ISO/IEC 14496-12
  aligned(8) class ProtectionSchemeInfoBox(fmt) extends Box('sinf') {
     OriginalFormatBox(fmt) original_format;
     SchemeTypeBox scheme_type_box; // optional
     SchemeInformationBox info; // optional
  }

### func (\*Box) [Append](./sinf.go#L22)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./sinf.go#L40)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[sinf.go](./sinf.go)
