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
   Header BoxHeader
   Boxes []Box
   OriginalFormat OriginalFormatBox
}

func (b *ProtectionSchemeInfoBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("*", "BoxType", head.BoxType())
      size := head.BoxPayload()
      switch head.BoxType() {
      case "schi", "schm":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      case "frma":
         b.OriginalFormat.Header = head
         err := b.OriginalFormat.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (b ProtectionSchemeInfoBox) Encode(w io.Writer) error {
   err := b.Header.Encode(w)
   if err != nil {
      return err
   }
   for _, value := range b.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   return b.OriginalFormat.Encode(w)
}
