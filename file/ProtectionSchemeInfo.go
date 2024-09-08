package sofia

import (
	"154.pages.dev/sofia/box"
	"io"
)

func (p *ProtectionSchemeInfo) read(r io.Reader, size int64) error {
	r = io.LimitReader(r, size)
	for {
		var head box.Header
		err := head.Read(r)
		switch err {
		case nil:
			switch head.Type.String() {
			case "frma":
				p.OriginalFormat.BoxHeader = head
				err := p.OriginalFormat.read(r)
				if err != nil {
					return err
				}
			case "schi":
				p.SchemeInformation.BoxHeader = head
				err := p.SchemeInformation.read(r)
				if err != nil {
					return err
				}
			case "schm": // Roku
				value := box.Box{BoxHeader: head}
				err := value.Read(r)
				if err != nil {
					return err
				}
				p.Boxes = append(p.Boxes, value)
			default:
				return box.Error{p.BoxHeader.Type, head.Type}
			}
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}

// ISO/IEC 14496-12
//
//	aligned(8) class ProtectionSchemeInfoBox(fmt) extends Box('sinf') {
//	   OriginalFormatBox(fmt) original_format;
//	   SchemeTypeBox scheme_type_box; // optional
//	   SchemeInformationBox info; // optional
//	}
type ProtectionSchemeInfo struct {
	BoxHeader         box.Header
	Boxes             []box.Box
	OriginalFormat    OriginalFormat
	SchemeInformation SchemeInformation
}

func (p ProtectionSchemeInfo) write(w io.Writer) error {
	err := p.BoxHeader.Write(w)
	if err != nil {
		return err
	}
	for _, value := range p.Boxes {
		err := value.Write(w)
		if err != nil {
			return err
		}
	}
	err = p.OriginalFormat.write(w)
	if err != nil {
		return err
	}
	return p.SchemeInformation.write(w)
}
