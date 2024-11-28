# Overview

package `frma`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./frma.go#L12)

```go
type Box struct {
  BoxHeader  sofia.BoxHeader
  DataFormat [4]uint8
}
```

ISO/IEC 14496-12
  aligned(8) class OriginalFormatBox(codingname) extends Box('frma') {
     unsigned int(32) data_format = codingname;
     // format of decrypted, encoded data (in case of protection)
     // or un-transformed sample entry (in case of restriction
     // and complete track information)
  }

### func (\*Box) [Append](./frma.go#L17)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./frma.go#L25)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[frma.go](./frma.go)
