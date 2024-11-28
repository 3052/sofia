# Overview

package `stsd`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
    - [func (b \*Box) SampleEntry() (\*sofia.SampleEntry, bool)](#func-box-sampleentry)
    - [func (b \*Box) Sinf() (\*sinf.Box, bool)](#func-box-sinf)
- [Source files](#source-files)

## Types

### type [Box](./stsd.go#L83)

```go
type Box struct {
  BoxHeader     sofia.BoxHeader
  FullBoxHeader sofia.FullBoxHeader
  EntryCount    uint32
  Box           []sofia.Box
  AudioSample   *enca.SampleEntry
  VisualSample  *encv.SampleEntry
}
```

ISO/IEC 14496-12
  aligned(8) class SampleDescriptionBox() extends FullBox('stsd', version, 0) {
     int i ;
     unsigned int(32) entry_count;
     for (i = 1 ; i <= entry_count ; i++){
        SampleEntry(); // an instance of a class derived from SampleEntry
     }
  }

### func (\*Box) [Append](./stsd.go#L92)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./stsd.go#L31)

```go
func (b *Box) Read(data []byte) error
```

### func (\*Box) [SampleEntry](./stsd.go#L11)

```go
func (b *Box) SampleEntry() (*sofia.SampleEntry, bool)
```

### func (\*Box) [Sinf](./stsd.go#L21)

```go
func (b *Box) Sinf() (*sinf.Box, bool)
```

## Source files

[stsd.go](./stsd.go)
