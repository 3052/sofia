package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//  aligned(8) class ProtectionSchemeInfoBox(fmt) extends Box('sinf') {
//     OriginalFormatBox(fmt) original_format;
//     SchemeTypeBox scheme_type_box; // optional
//     SchemeInformationBox info; // optional
//  }
type ProtectionSchemeInfo struct {
   BoxHeader      BoxHeader
   Boxes          []Box
   OriginalFormat OriginalFormat
   SchemeInformation SchemeInformation
}

func (p *ProtectionSchemeInfo) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.debug() {
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
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         p.Boxes = append(p.Boxes, b)
      default:
         return errors.New("ProtectionSchemeInfo.read")
      }
   }
}

func (p ProtectionSchemeInfo) write(w io.Writer) error {
   err := p.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, b := range p.Boxes {
      err := b.write(w)
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
