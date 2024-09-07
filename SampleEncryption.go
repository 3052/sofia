package sofia

import (
	"154.pages.dev/sofia/box"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"io"
)

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
	BoxHeader     box.Header
	FullBoxHeader FullBoxHeader
	SampleCount   uint32
	Samples       []EncryptionSample
}

type EncryptionSample struct {
	InitializationVector uint64
	SubsampleCount       uint16
	Subsamples           []Subsample
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

func (e *EncryptionSample) read(r io.Reader, box *SampleEncryption) error {
	err := binary.Read(r, binary.BigEndian, &e.InitializationVector)
	if err != nil {
		return err
	}
	if box.senc_use_subsamples() {
		err := binary.Read(r, binary.BigEndian, &e.SubsampleCount)
		if err != nil {
			return err
		}
		e.Subsamples = make([]Subsample, e.SubsampleCount)
		for i, sample := range e.Subsamples {
			err := sample.read(r)
			if err != nil {
				return err
			}
			e.Subsamples[i] = sample
		}
	}
	return nil
}

func (e EncryptionSample) write(w io.Writer, box SampleEncryption) error {
	err := binary.Write(w, binary.BigEndian, e.InitializationVector)
	if err != nil {
		return err
	}
	if box.senc_use_subsamples() {
		err := binary.Write(w, binary.BigEndian, e.SubsampleCount)
		if err != nil {
			return err
		}
		for _, sample := range e.Subsamples {
			err := sample.write(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SampleEncryption) read(r io.Reader) error {
	err := s.FullBoxHeader.read(r)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &s.SampleCount)
	if err != nil {
		return err
	}
	s.Samples = make([]EncryptionSample, s.SampleCount)
	for i, sample := range s.Samples {
		err := sample.read(r, s)
		if err != nil {
			return err
		}
		s.Samples[i] = sample
	}
	return nil
}

// senc_use_subsamples: flag mask is 0x000002.
func (s SampleEncryption) senc_use_subsamples() bool {
	return s.FullBoxHeader.get_flags()&2 >= 1
}

func (s SampleEncryption) write(w io.Writer) error {
	err := s.BoxHeader.write(w)
	if err != nil {
		return err
	}
	err = s.FullBoxHeader.write(w)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, s.SampleCount)
	if err != nil {
		return err
	}
	for _, sample := range s.Samples {
		err := sample.write(w, s)
		if err != nil {
			return err
		}
	}
	return nil
}

type Subsample struct {
	BytesOfClearData     uint16
	BytesOfProtectedData uint32
}

func (s *Subsample) read(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, s)
}

func (s Subsample) write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, s)
}
