package sofia

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"io"
)

type EncryptionSample struct {
	InitializationVector uint64
	SubsampleCount       uint16
	Subsamples           []Subsample
}

func (e *EncryptionSample) Decode(r io.Reader, b *SampleEncryption) error {
	err := binary.Read(r, binary.BigEndian, &e.InitializationVector)
	if err != nil {
		return err
	}
	if b.SencUseSubsamples() {
		err := binary.Read(r, binary.BigEndian, &e.SubsampleCount)
		if err != nil {
			return err
		}
		e.Subsamples = make([]Subsample, e.SubsampleCount)
		for i, sample := range e.Subsamples {
			err := sample.Decode(r)
			if err != nil {
				return err
			}
			e.Subsamples[i] = sample
		}
	}
	return nil
}

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func (e EncryptionSample) DecryptCenc(sample, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	var iv [16]byte
	binary.BigEndian.PutUint64(iv[:], e.InitializationVector)
	stream := cipher.NewCTR(block, iv[:])
	if len(e.Subsamples) >= 1 {
		var pos uint32
		for _, ss := range e.Subsamples {
			nrClear := uint32(ss.BytesOfClearData)
			if nrClear >= 1 {
				pos += nrClear
			}
			nrEnc := ss.BytesOfProtectedData
			if nrEnc >= 1 {
				stream.XORKeyStream(sample[pos:pos+nrEnc], sample[pos:pos+nrEnc])
				pos += nrEnc
			}
		}
	} else {
		stream.XORKeyStream(sample, sample)
	}
	return nil
}

func (e EncryptionSample) Encode(w io.Writer, b SampleEncryption) error {
	err := binary.Write(w, binary.BigEndian, e.InitializationVector)
	if err != nil {
		return err
	}
	if b.SencUseSubsamples() {
		err := binary.Write(w, binary.BigEndian, e.SubsampleCount)
		if err != nil {
			return err
		}
		for _, sample := range e.Subsamples {
			err := sample.Encode(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ISO/IEC 23001-7
//
// if the version of the SampleEncryptionBox is 0 and the flag
// senc_use_subsamples is set, UseSubSampleEncryption is set to 1
//
//	aligned(8) class SampleEncryptionBox extends FullBox(
//	   'senc', version, flags
//	) {
//	   unsigned int(32) sample_count;
//	   {
//	      unsigned int(Per_Sample_IV_Size*8) InitializationVector;
//	      if (UseSubSampleEncryption) {
//	         unsigned int(16) subsample_count;
//	         {
//	            unsigned int(16) BytesOfClearData;
//	            unsigned int(32) BytesOfProtectedData;
//	         } [subsample_count ]
//	      }
//	   }[ sample_count ]
//	}
type SampleEncryption struct {
	BoxHeader     BoxHeader
	FullBoxHeader FullBoxHeader
	SampleCount   uint32
	Samples       []EncryptionSample
}

func (b *SampleEncryption) Decode(r io.Reader) error {
	err := b.FullBoxHeader.read(r)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &b.SampleCount)
	if err != nil {
		return err
	}
	b.Samples = make([]EncryptionSample, b.SampleCount)
	for i, sample := range b.Samples {
		err := sample.Decode(r, b)
		if err != nil {
			return err
		}
		b.Samples[i] = sample
	}
	return nil
}

func (b SampleEncryption) Encode(w io.Writer) error {
	err := b.BoxHeader.write(w)
	if err != nil {
		return err
	}
	if err := b.FullBoxHeader.write(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, b.SampleCount); err != nil {
		return err
	}
	for _, sample := range b.Samples {
		err := sample.Encode(w, b)
		if err != nil {
			return err
		}
	}
	return nil
}

// senc_use_subsamples: flag mask is 0x000002.
func (b SampleEncryption) SencUseSubsamples() bool {
	return b.FullBoxHeader.get_flags()&2 >= 1
}

type Subsample struct {
	BytesOfClearData     uint16
	BytesOfProtectedData uint32
}

func (s *Subsample) Decode(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, s)
}

func (s Subsample) Encode(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, s)
}
