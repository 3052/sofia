# Overview

package `container`

## Index

- [Types](#types)
  - [type File](#type-file)
    - [func (f \*File) Append(data []byte) ([]byte, error)](#func-file-append)
    - [func (f \*File) GetMoov() (\*moov.Box, bool)](#func-file-getmoov)
    - [func (f \*File) Read(data []byte) error](#func-file-read)
- [Source files](#source-files)

## Types

### type [File](./container.go#L56)

```go
type File struct {
  Box  []sofia.Box
  Mdat *mdat.Box
  Moof *moof.Box
  Moov *moov.Box
  Sidx *sidx.Box
}
```

ISO/IEC 14496-12

### func (\*File) [Append](./container.go#L64)

```go
func (f *File) Append(data []byte) ([]byte, error)
```

### func (\*File) [GetMoov](./container.go#L100)

```go
func (f *File) GetMoov() (*moov.Box, bool)
```

### func (\*File) [Read](./container.go#L12)

```go
func (f *File) Read(data []byte) error
```

## Source files

[container.go](./container.go)
