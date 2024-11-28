# Overview

package `senc`

## Index

- [Types](#types)
  - [type Box](#type-box)
    - [func (b \*Box) Append(data []byte) ([]byte, error)](#func-box-append)
    - [func (b \*Box) Read(data []byte) error](#func-box-read)
  - [type Sample](#type-sample)
    - [func (s \*Sample) Append(data []byte) ([]byte, error)](#func-sample-append)
    - [func (s \*Sample) Decode(data []byte) (int, error)](#func-sample-decode)
    - [func (s \*Sample) DecryptCenc(text, key []byte) error](#func-sample-decryptcenc)
  - [type Subsample](#type-subsample)
    - [func (s Subsample) Append(data []byte) ([]byte, error)](#func-subsample-append)
    - [func (s \*Subsample) Decode(data []byte) (int, error)](#func-subsample-decode)
- [Source files](#source-files)

## Types

### type [Box](./senc.go#L31)

```go
type Box struct {
  BoxHeader     sofia.BoxHeader
  FullBoxHeader sofia.FullBoxHeader
  SampleCount   uint32
  Sample        []Sample
}
```

ISO/IEC 23001-7

if the version of the SampleEncryptionBox is 0 and the flag
senc_use_subsamples is set, UseSubSampleEncryption is set to 1

  aligned(8) class SampleEncryptionBox extends FullBox(
     'senc', version, flags
  ) {
     unsigned int(32) sample_count;
     {
        unsigned int(Per_Sample_IV_Size*8) InitializationVector;
        if (UseSubSampleEncryption) {
           unsigned int(16) subsample_count;
           {
              unsigned int(16) BytesOfClearData;
              unsigned int(32) BytesOfProtectedData;
           } [subsample_count ]
        }
     }[ sample_count ]
  }

### func (\*Box) [Append](./senc.go#L38)

```go
func (b *Box) Append(data []byte) ([]byte, error)
```

### func (\*Box) [Read](./senc.go#L121)

```go
func (b *Box) Read(data []byte) error
```

### type [Sample](./senc.go#L114)

```go
type Sample struct {
  InitializationVector [8]uint8
  SubsampleCount       uint16
  Subsample            []Subsample
  // contains filtered or unexported fields
}
```

### func (\*Sample) [Append](./senc.go#L71)

```go
func (s *Sample) Append(data []byte) ([]byte, error)
```

### func (\*Sample) [Decode](./senc.go#L147)

```go
func (s *Sample) Decode(data []byte) (int, error)
```

### func (\*Sample) [DecryptCenc](./senc.go#L87)

```go
func (s *Sample) DecryptCenc(text, key []byte) error
```

github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101

### type [Subsample](./senc.go#L66)

```go
type Subsample struct {
  BytesOfClearData     uint16
  BytesOfProtectedData uint32
}
```

### func (Subsample) [Append](./senc.go#L62)

```go
func (s Subsample) Append(data []byte) ([]byte, error)
```

### func (\*Subsample) [Decode](./senc.go#L168)

```go
func (s *Subsample) Decode(data []byte) (int, error)
```

## Source files

[senc.go](./senc.go)
