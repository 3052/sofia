package file

import (
	"154.pages.dev/sofia"
	"encoding/binary"
	"io"
)

// ISO/IEC 23001-7
//
//	aligned(8) class TrackEncryptionBox extends FullBox('tenc', version, flags=0) {
//	   unsigned int(8) reserved = 0;
//	   if (version==0) {
//	      unsigned int(8) reserved = 0;
//	   } else { // version is 1 or greater
//	      unsigned int(4) default_crypt_byte_block;
//	      unsigned int(4) default_skip_byte_block;
//	   }
//	   unsigned int(8) default_isProtected;
//	   unsigned int(8) default_Per_Sample_IV_Size;
//	   unsigned int(8)[16] default_KID;
//	   if (default_isProtected ==1 && default_Per_Sample_IV_Size == 0) {
//	      unsigned int(8) default_constant_IV_size;
//	      unsigned int(8)[default_constant_IV_size] default_constant_IV;
//	   }
//	}
type TrackEncryption struct {
	BoxHeader     sofia.BoxHeader
	FullBoxHeader sofia.FullBoxHeader
	Extends       struct {
		_                      uint8
		_                      uint8
		DefaultIsProtected     uint8
		DefaultPerSampleIvSize uint8
	}
	DefaultKid sofia.UUID
}

func (t *TrackEncryption) read(r io.Reader) error {
	err := t.BoxHeader.Read(r)
	if err != nil {
		return err
	}
	err = t.FullBoxHeader.Read(r)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &t.Extends)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(r, t.DefaultKid[:])
	if err != nil {
		return err
	}
	return nil
}

func (t TrackEncryption) write(w io.Writer) error {
	err := t.BoxHeader.Write(w)
	if err != nil {
		return err
	}
	err = t.FullBoxHeader.Write(w)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, t.Extends)
	if err != nil {
		return err
	}
	_, err = w.Write(t.DefaultKid[:])
	if err != nil {
		return err
	}
	return nil
}
