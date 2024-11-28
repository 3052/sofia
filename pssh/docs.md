# Overview

package `pssh`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
    - [func (b \*Box) Widevine() bool](#func-box-widevine)
- [Source files](#source-files)

## Types

### type [Box](./pssh.go#L22)

```go
type Box struct {
  BoxHeader     sofia.BoxHeader
  FullBoxHeader sofia.FullBoxHeader
  SystemId      sofia.Uuid
  KidCount      uint32
  Kid           []sofia.Uuid
  DataSize      uint32
  Data          []uint8
}
```

ISO/IEC 23001-7
  aligned(8) class ProtectionSystemSpecificHeaderBox extends FullBox(
     'pssh', version, flags=0,
  ) {
     unsigned int(8)[16] SystemID;
     if (version > 0) {
        unsigned int(32) KID_count;
        {
           unsigned int(8)[16] KID;
        } [KID_count];
     }
     unsigned int(32) DataSize;
     unsigned int(8)[DataSize] Data;
  }

### func (\*Box) [Append](./pssh.go#L37)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./pssh.go#L58)

```go
func (b *Box) Read(data []byte) error
```

### func (\*Box) [Widevine](./pssh.go#L33)

```go
func (b *Box) Widevine() bool
```

dashif.org/identifiers/content_protection

## Source files

[pssh.go](./pssh.go)
