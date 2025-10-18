// File: pssh_box.go
package mp4parser

import "bytes"

// WidevineSystemID is the specific UUID for the Widevine DRM system.
var WidevineSystemID = []byte{0xed, 0xef, 0x8b, 0xa9, 0x79, 0xd6, 0x4a, 0xce, 0xa3, 0xc8, 0x27, 0xdc, 0xd5, 0x1d, 0x21, 0xed}

// PsshBox (Protection System Specific Header Box)
type PsshBox struct {
	FullBox
	SystemID []byte // 16 bytes
	KIDs     [][]byte // Optional, for Widevine v1+
	Data     []byte
}

// ParsePsshBox parses the PsshBox from its content slice.
func ParsePsshBox(data []byte) (*PsshBox, error) {
	b := &PsshBox{}
	offset, err := b.FullBox.Parse(data, 0)
	if err != nil {
		return nil, err
	}

	if offset+16 > len(data) {
		return nil, ErrUnexpectedEOF
	}
	b.SystemID = data[offset : offset+16]
	offset += 16

	isWidevine := bytes.Equal(b.SystemID, WidevineSystemID)

	// Widevine pssh boxes have a more complex structure
	if isWidevine {
		if b.Version > 0 {
			var kidCount uint32
			kidCount, offset, err = readUint32(data, offset)
			if err != nil {
				return nil, err
			}
			for i := 0; i < int(kidCount); i++ {
				if offset+16 > len(data) {
					return nil, ErrUnexpectedEOF
				}
				b.KIDs = append(b.KIDs, data[offset:offset+16])
				offset += 16
			}
		}
		var dataSize uint32
		dataSize, offset, err = readUint32(data, offset)
		if err != nil {
			return nil, err
		}
		if offset+int(dataSize) > len(data) {
			return nil, ErrUnexpectedEOF
		}
		b.Data = data[offset : offset+int(dataSize)]
	} else {
		// For other systems like PlayReady, the rest of the box is data.
		b.Data = data[offset:]
	}
	return b, nil
}

// Size calculates the total byte size of the PsshBox.
func (b *PsshBox) Size() uint64 {
	size := uint64(8) // Header
	size += b.FullBox.Size()
	size += 16 // SystemID

	isWidevine := bytes.Equal(b.SystemID, WidevineSystemID)
	if isWidevine {
		if b.Version > 0 {
			size += 4 // KIDCount
			size += uint64(len(b.KIDs) * 16)
		}
		size += 4 // DataSize
	}

	size += uint64(len(b.Data))
	return size
}

// Format serializes the PsshBox into the destination slice.
func (b *PsshBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "pssh")
	offset = b.FullBox.Format(dst, offset)
	offset = writeBytes(dst, offset, b.SystemID)

	isWidevine := bytes.Equal(b.SystemID, WidevineSystemID)
	if isWidevine {
		if b.Version > 0 {
			offset = writeUint32(dst, offset, uint32(len(b.KIDs)))
			for _, kid := range b.KIDs {
				offset = writeBytes(dst, offset, kid)
			}
		}
		offset = writeUint32(dst, offset, uint32(len(b.Data)))
	}

	offset = writeBytes(dst, offset, b.Data)
	return offset
}