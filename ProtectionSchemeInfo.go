package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: Protected Sample Entry
//  aligned(8) class ProtectionSchemeInfoBox(fmt) extends Box('sinf') {
//     OriginalFormatBox(fmt) original_format;
//     SchemeTypeBox scheme_type_box; // optional
//     SchemeInformationBox info; // optional
//  }
type ProtectionSchemeInfoBox struct {
   BoxHeader BoxHeader
   Boxes []Box
   OriginalFormat OriginalFormatBox
}

func (p *ProtectionSchemeInfoBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("*", "BoxType", head.BoxType())
      r := head.Reader(r)
      switch head.BoxType() {
      case "schi", "schm":
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
         return errors.New("BoxType")
      }
   }
}

func (p ProtectionSchemeInfoBox) Encode(w io.Writer) error {
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
