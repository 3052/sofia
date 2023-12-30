package sofia

import "io"

// Container: Protected Sample Entry
//  aligned(8) class ProtectionSchemeInfoBox(fmt) extends Box('sinf') {
//     OriginalFormatBox(fmt) original_format;
//     SchemeTypeBox scheme_type_box; // optional
//     SchemeInformationBox info; // optional
//  }
type ProtectionSchemeInfoBox struct {
   Header BoxHeader
   Payload []byte
}

func (b *ProtectionSchemeInfoBox) Decode(r io.Reader) error {
   return nil
}
