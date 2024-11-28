# Overview

package `tenc`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
- [Source files](#source-files)

## Types

### type [Box](./tenc.go#L45)

```go
type Box struct {
  BoxHeader sofia.BoxHeader
  Fixed     struct {
    FullBoxHeader          sofia.FullBoxHeader
    Reserved               uint8
    ByteBlock              uint8
    DefaultIsProtected     uint8
    DefaultPerSampleIvSize uint8
    DefaultKid             sofia.Uuid
  }
  DefaultConstantIvSize uint8
  DefaultConstantIv     []uint8
}
```

ISO/IEC 23001-7
  aligned(8) class TrackEncryptionBox extends FullBox('tenc', version, flags=0) {
     unsigned int(8) reserved = 0;
     if (version==0) {
        unsigned int(8) reserved = 0;
     } else { // version is 1 or greater
        unsigned int(4) default_crypt_byte_block;
        unsigned int(4) default_skip_byte_block;
     }
     unsigned int(8) default_isProtected;
     unsigned int(8) default_Per_Sample_IV_Size;
     unsigned int(8)[16] default_KID;
     if (default_isProtected ==1 && default_Per_Sample_IV_Size == 0) {
        unsigned int(8) default_constant_IV_size;
        unsigned int(8)[default_constant_IV_size] default_constant_IV;
     }
  }

### func (\*Box) [Append](./tenc.go#L59)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./tenc.go#L8)

```go
func (b *Box) Read(data []byte) error
```

## Source files

[tenc.go](./tenc.go)
