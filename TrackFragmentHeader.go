package sofia

import (
	"154.pages.dev/sofia/box"
	"encoding/binary"
	"io"
)

// ISO/IEC 14496-12
//
//	aligned(8) class TrackFragmentHeaderBox extends FullBox(
//	   'tfhd', 0, tf_flags
//	) {
//	   unsigned int(32) track_ID;
//	   // all the following are optional fields
//	   // their presence is indicated by bits in the tf_flags
//	   unsigned int(64) base_data_offset; // ASSUME NOT PRESENT
//	   unsigned int(32) sample_description_index;
//	   unsigned int(32) default_sample_duration;
//	   unsigned int(32) default_sample_size;
//	   unsigned int(32) default_sample_flags;
//	}
type TrackFragmentHeader struct {
	BoxHeader              box.Header
	FullBoxHeader          box.FullBoxHeader
	TrackId                uint32
	SampleDescriptionIndex uint32
	DefaultSampleDuration  uint32
	DefaultSampleSize      uint32
	DefaultSampleFlags     uint32
}

func (t TrackFragmentHeader) write(w io.Writer) error {
	err := t.BoxHeader.Write(w)
	if err != nil {
		return err
	}
	err = t.FullBoxHeader.Write(w)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, t.TrackId)
	if err != nil {
		return err
	}
	if t.sample_description_index_present() {
		err := binary.Write(w, binary.BigEndian, t.SampleDescriptionIndex)
		if err != nil {
			return err
		}
	}
	if t.default_sample_duration_present() {
		err := binary.Write(w, binary.BigEndian, t.DefaultSampleDuration)
		if err != nil {
			return err
		}
	}
	if t.default_sample_size_present() {
		err := binary.Write(w, binary.BigEndian, t.DefaultSampleSize)
		if err != nil {
			return err
		}
	}
	if t.default_sample_flags_present() {
		err := binary.Write(w, binary.BigEndian, t.DefaultSampleFlags)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TrackFragmentHeader) read(r io.Reader) error {
	err := t.FullBoxHeader.Read(r)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &t.TrackId)
	if err != nil {
		return err
	}
	if t.sample_description_index_present() {
		err := binary.Read(r, binary.BigEndian, &t.SampleDescriptionIndex)
		if err != nil {
			return err
		}
	}
	if t.default_sample_duration_present() {
		err := binary.Read(r, binary.BigEndian, &t.DefaultSampleDuration)
		if err != nil {
			return err
		}
	}
	if t.default_sample_size_present() {
		err := binary.Read(r, binary.BigEndian, &t.DefaultSampleSize)
		if err != nil {
			return err
		}
	}
	if t.default_sample_flags_present() {
		err := binary.Read(r, binary.BigEndian, &t.DefaultSampleFlags)
		if err != nil {
			return err
		}
	}
	return nil
}

// 0x000002 sample-description-index-present
func (t TrackFragmentHeader) sample_description_index_present() bool {
	return t.FullBoxHeader.GetFlags()&0x2 >= 1
}

// 0x000008 default-sample-duration-present
func (t TrackFragmentHeader) default_sample_duration_present() bool {
	return t.FullBoxHeader.GetFlags()&0x8 >= 1
}

// 0x000010 default-sample-size-present
func (t TrackFragmentHeader) default_sample_size_present() bool {
	return t.FullBoxHeader.GetFlags()&0x10 >= 1
}

// 0x000020 default-sample-flags-present
func (t TrackFragmentHeader) default_sample_flags_present() bool {
	return t.FullBoxHeader.GetFlags()&0x20 >= 1
}
