# Overview

package `mdat`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Data(track \*traf.Box) [][]byte](#func-box-data)
- [Source files](#source-files)

## Types

### type [Box](./mdat.go#L12)

```go
type Box struct {
  Box sofia.Box
}
```

ISO/IEC 14496-12
  aligned(8) class MediaDataBox extends Box('mdat') {
     bit(8) data[];
  }

### func (\*Box) [Data](./mdat.go#L17)

```go
func (b *Box) Data(track *traf.Box) [][]byte
```

BE CAREFUL WITH THE RECEIVER

## Source files

[mdat.go](./mdat.go)
