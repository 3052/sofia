package box

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"strconv"
)

type Error struct {
	Container Type
	Box       Type
}

func (e Error) Error() string {
	b := []byte("container:")
	b = strconv.AppendQuote(b, e.Container.String())
	b = append(b, " box type:"...)
	b = strconv.AppendQuote(b, e.Box.String())
	return string(b)
}

func (b *Header) GetSize() (int, int64) {
	size := binary.Size(b.Size)
	size += binary.Size(b.Type)
	if b.Type.String() == "uuid" {
		size += binary.Size(b.UserType)
	}
	return size, int64(b.Size) - int64(size)
}

// ISO/IEC 14496-12
//
//	aligned(8) class BoxHeader (
//	   unsigned int(32) boxtype,
//	   optional unsigned int(8)[16] extended_type
//	) {
//	   unsigned int(32) size;
//	   unsigned int(32) type = boxtype;
//	   if (size==1) {
//	      unsigned int(64) largesize;
//	   } else if (size==0) {
//	      // box extends to end of file
//	   }
//	   if (boxtype=='uuid') {
//	      unsigned int(8)[16] usertype = extended_type;
//	   }
//	}
type Header struct {
	Size     uint32
	Type     Type
	UserType UUID
}

func (b *Header) Read(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &b.Size)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(r, b.Type[:])
	if err != nil {
		return err
	}
	if b.Type.String() == "uuid" {
		_, err := io.ReadFull(r, b.UserType[:])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Header) write(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, b.Size)
	if err != nil {
		return err
	}
	_, err = w.Write(b.Type[:])
	if err != nil {
		return err
	}
	if b.Type.String() == "uuid" {
		_, err := w.Write(b.UserType[:])
		if err != nil {
			return err
		}
	}
	return nil
}

// ISO/IEC 14496-12
//
//	aligned(8) class FullBoxHeader(unsigned int(8) v, bit(24) f) {
//	   unsigned int(8) version = v;
//	   bit(24) flags = f;
//	}
type FullBoxHeader struct {
	Version uint8
	Flags   [3]byte
}

type UUID [16]uint8

func (u UUID) String() string {
	return hex.EncodeToString(u[:])
}

func (t Type) String() string {
	return string(t[:])
}

type Type [4]uint8

// ISO/IEC 14496-12
//
//	aligned(8) class Box (
//	   unsigned int(32) boxtype,
//	   optional unsigned int(8)[16] extended_type
//	) {
//	   BoxHeader(boxtype, extended_type);
//	   // the remaining bytes are the BoxPayload
//	}
type Box struct {
	BoxHeader Header
	Payload   []byte
}

func (b *Box) Read(r io.Reader) error {
	_, size := b.BoxHeader.GetSize()
	b.Payload = make([]byte, size)
	_, err := io.ReadFull(r, b.Payload)
	if err != nil {
		return err
	}
	return nil
}

func (b *Box) Write(w io.Writer) error {
	err := b.BoxHeader.write(w)
	if err != nil {
		return err
	}
	_, err = w.Write(b.Payload)
	if err != nil {
		return err
	}
	return nil
}