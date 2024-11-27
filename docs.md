# Overview

package `sofia`

## Index

- [Types](#types)
  - [type Appender](#type-appender)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
  - [type BoxHeader](#type-boxheader)
    - [func (b \*BoxHeader) Append(data []byte) ([]byte, error)](#func-boxheader-append)
    - [func (b \*BoxHeader) Decode(data []byte) (int, error)](#func-boxheader-decode)
    - [func (b \*BoxHeader) GetSize() int](#func-boxheader-getsize)
  - [type Decoder](#type-decoder)
  - [type Error](#type-error)
    - [func (e \*Error) Error() string](#func-error-error)
  - [type FullBoxHeader](#type-fullboxheader)
    - [func (f \*FullBoxHeader) Append(data []byte) ([]byte, error)](#func-fullboxheader-append)
    - [func (f \*FullBoxHeader) Decode(data []byte) (int, error)](#func-fullboxheader-decode)
    - [func (f \*FullBoxHeader) GetFlags() uint32](#func-fullboxheader-getflags)
  - [type Reader](#type-reader)
  - [type SampleEntry](#type-sampleentry)
    - [func (s \*SampleEntry) Append(data []byte) ([]byte, error)](#func-sampleentry-append)
    - [func (s \*SampleEntry) Decode(data []byte) (int, error)](#func-sampleentry-decode)
  - [type SizeGetter](#type-sizegetter)
  - [type Type](#type-type)
    - [func (t Type) String() string](#func-type-string)
  - [type Uuid](#type-uuid)
    - [func (u Uuid) String() string](#func-uuid-string)
- [Source files](#source-files)

## Types

### type [Appender](./sofia.go#L50)

```go
type Appender interface {
  Append([]byte) ([]byte, error)
}
```

### type [Box](./sofia.go#L45)

```go
type Box struct {
  BoxHeader BoxHeader
  Payload   []byte
}
```

ISO/IEC 14496-12
  aligned(8) class Box (
     unsigned int(32) boxtype,
     optional unsigned int(8)[16] extended_type
  ) {
     BoxHeader(boxtype, extended_type);
     // the remaining bytes are the BoxPayload
  }

### func (\*Box) [Append](./sofia.go#L63)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./sofia.go#L54)

```go
func (b *Box) Read(data []byte) error
```

### type [BoxHeader](./sofia.go#L31)

```go
type BoxHeader struct {
  Size     uint32
  Type     Type
  UserType *Uuid
}
```

ISO/IEC 14496-12
  aligned(8) class BoxHeader (
     unsigned int(32) boxtype,
     optional unsigned int(8)[16] extended_type
  ) {
     unsigned int(32) size;
     unsigned int(32) type = boxtype;
     if (size==1) {
        unsigned int(64) largesize;
     } else if (size==0) {
        // box extends to end of file
     }
     if (boxtype=='uuid') {
        unsigned int(8)[16] usertype = extended_type;
     }
  }

### func (\*BoxHeader) [Append](./sofia.go#L80)

```go
func (b *BoxHeader) Append(data []byte) ([]byte, error)
```

### func (\*BoxHeader) [Decode](./sofia.go#L89)

```go
func (b *BoxHeader) Decode(data []byte) (int, error)
```

### func (\*BoxHeader) [GetSize](./sofia.go#L71)

```go
func (b *BoxHeader) GetSize() int
```

### type [Decoder](./sofia.go#L102)

```go
type Decoder interface {
  Decode([]byte) (int, error)
}
```

### type [Error](./sofia.go#L106)

```go
type Error struct {
  Container BoxHeader
  Box       BoxHeader
}
```

### func (\*Error) [Error](./sofia.go#L111)

```go
func (e *Error) Error() string
```

### type [FullBoxHeader](./sofia.go#L124)

```go
type FullBoxHeader struct {
  Version uint8
  Flags   [3]byte
}
```

ISO/IEC 14496-12
  aligned(8) class FullBoxHeader(unsigned int(8) v, bit(24) f) {
     unsigned int(8) version = v;
     bit(24) flags = f;
  }

### func (\*FullBoxHeader) [Append](./sofia.go#L135)

```go
func (f *FullBoxHeader) Append(data []byte) ([]byte, error)
```

### func (\*FullBoxHeader) [Decode](./sofia.go#L139)

```go
func (f *FullBoxHeader) Decode(data []byte) (int, error)
```

### func (\*FullBoxHeader) [GetFlags](./sofia.go#L129)

```go
func (f *FullBoxHeader) GetFlags() uint32
```

### type [Reader](./sofia.go#L143)

```go
type Reader interface {
  Read([]byte) error
}
```

### type [SampleEntry](./sofia.go#L163)

```go
type SampleEntry struct {
  BoxHeader          BoxHeader
  Reserved           [6]uint8
  DataReferenceIndex uint16
}
```

ISO/IEC 14496-12
  aligned(8) abstract class SampleEntry(
     unsigned int(32) format
  ) extends Box(format) {
     const unsigned int(8)[6] reserved = 0;
     unsigned int(16) data_reference_index;
  }

### func (\*SampleEntry) [Append](./sofia.go#L169)

```go
func (s *SampleEntry) Append(data []byte) ([]byte, error)
```

### func (\*SampleEntry) [Decode](./sofia.go#L147)

```go
func (s *SampleEntry) Decode(data []byte) (int, error)
```

### type [SizeGetter](./sofia.go#L178)

```go
type SizeGetter interface {
  GetSize() int
}
```

### type [Type](./sofia.go#L186)

```go
type Type [4]uint8
```

### func (Type) [String](./sofia.go#L182)

```go
func (t Type) String() string
```

### type [Uuid](./sofia.go#L13)

```go
type Uuid [16]uint8
```

### func (Uuid) [String](./sofia.go#L9)

```go
func (u Uuid) String() string
```

## Source files

[sofia.go](./sofia.go)
