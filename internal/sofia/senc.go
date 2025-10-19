// File: senc.go
package mp4parser

import "log" // Import the log package

type SencBox struct {
	FullBox
	SampleCount           uint32
	InitializationVectors []InitializationVector
}

// ... (InitializationVector and Subsample structs are unchanged)
type InitializationVector struct {
	IV         []byte
	Subsamples []Subsample
}
type Subsample struct {
	BytesOfClearData     uint16
	BytesOfProtectedData uint32
}

func ParseSencContent(data []byte, perSampleIVSize uint8, defaultConstantIV []byte) (*SencBox, error) {
	b := &SencBox{}
	// LOGGING: Log the input conditions
	log.Printf("[ParseSencContent] Starting parse. Data length: %d, perSampleIVSize: %d", len(data), perSampleIVSize)

	offset, err := b.FullBox.Parse(data, 0)
	if err != nil {
		log.Printf("[ParseSencContent] ERROR parsing FullBox: %v", err)
		return nil, err
	}

	b.SampleCount, offset, err = readUint32(data, offset)
	if err != nil {
		log.Printf("[ParseSencContent] ERROR reading SampleCount at offset %d: %v", offset, err)
		return nil, err
	}

	// LOGGING: Log the critical SampleCount value
	log.Printf("[ParseSencContent] Parsed SampleCount: %d. Current offset: %d", b.SampleCount, offset)

	flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
	hasSubsamples := (flags & 0x000002) != 0
	log.Printf("[ParseSencContent] hasSubsamples flag is: %v", hasSubsamples)

	b.InitializationVectors = make([]InitializationVector, b.SampleCount)
	for i := 0; i < int(b.SampleCount); i++ {
		// LOGGING: Log progress inside the main loop
		log.Printf("[ParseSencContent] Processing sample %d/%d. Current offset: %d", i+1, b.SampleCount, offset)

		iv := InitializationVector{}
		if perSampleIVSize > 0 {
			ivSize := int(perSampleIVSize)
			if offset+ivSize > len(data) {
				log.Printf("[ParseSencContent] ERROR: Not enough data for IV. Need %d bytes at offset %d, but only %d bytes available.", ivSize, offset, len(data)-offset)
				return nil, ErrUnexpectedEOF
			}
			iv.IV = data[offset : offset+ivSize]
			offset += ivSize
		} else {
			iv.IV = defaultConstantIV
		}

		if hasSubsamples {
			var subsampleCount uint16
			subsampleCount, offset, err = readUint16(data, offset)
			if err != nil {
				log.Printf("[ParseSencContent] ERROR reading subsampleCount for sample %d at offset %d: %v", i+1, offset, err)
				return nil, err
			}
			iv.Subsamples = make([]Subsample, subsampleCount)
			for j := 0; j < int(subsampleCount); j++ {
				// LOGGING: Log progress in the subsample loop
				log.Printf("[ParseSencContent]   - Reading subsample %d/%d. Offset: %d", j+1, subsampleCount, offset)
				var clearData uint16
				clearData, offset, err = readUint16(data, offset)
				if err != nil {
					log.Printf("[ParseSencContent] ERROR reading BytesOfClearData for sample %d, subsample %d: %v", i+1, j+1, err)
					return nil, err
				}
				var protectedData uint32
				protectedData, offset, err = readUint32(data, offset)
				if err != nil {
					log.Printf("[ParseSencContent] ERROR reading BytesOfProtectedData for sample %d, subsample %d: %v", i+1, j+1, err)
					return nil, err
				}
				iv.Subsamples[j] = Subsample{
					BytesOfClearData:     clearData,
					BytesOfProtectedData: protectedData,
				}
			}
		}
		b.InitializationVectors[i] = iv
	}
	log.Printf("[ParseSencContent] Successfully finished parsing. Final offset: %d", offset)
	return b, nil
}
