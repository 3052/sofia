package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//  aligned(8) class ProtectionSchemeInfoBox(fmt) extends Box('sinf') {
//     OriginalFormatBox(fmt) original_format;
//     SchemeTypeBox scheme_type_box; // optional
//     SchemeInformationBox info; // optional
//  }
type ProtectionSchemeInfo struct {
   BoxHeader BoxHeader
   Boxes []Box
   OriginalFormat OriginalFormat
}

func (p *ProtectionSchemeInfo) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("BoxHeader", "type", head.BoxType())
      r := head.BoxPayload(r)
      switch head.BoxType() {
      case "schi", // Roku
      "schm": // Roku
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         p.Boxes = append(p.Boxes, b)
      case "frma":
         p.OriginalFormat.BoxHeader = head
         err := p.OriginalFormat.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("ProtectionSchemeInfo.Decode")
      }
   }
}

func (p ProtectionSchemeInfo) Encode(w io.Writer) error {
   err := p.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   for _, b := range p.Boxes {
      err := b.Encode(w)
      if err != nil {
         return err
      }
   }
   return p.OriginalFormat.Encode(w)
}
