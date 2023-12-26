package sofia

import (
   "fmt"
   "io"
)

func (b *TrackFragmentBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.BoxPayload()
      switch head.Type() {
      case "senc":
         b.Senc.BoxHeader = head
         err := b.Senc.Decode(r)
         if err != nil {
            return err
         }
      case "trun":
         b.Trun.BoxHeader = head
         err := b.Trun.Decode(r)
         if err != nil {
            return err
         }
      case "saio", "saiz", "sbgp", "sgpd", "tfdt", "tfhd", "uuid":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      default:
         return fmt.Errorf("traf %q", head.RawType)
      }
   }
}

func (b TrackFragmentBox) Encode(w io.Writer) error {
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
   if err := b.Trun.Encode(w); err != nil {
      return err
   }
   return b.Senc.Encode(w)
}

// 8.8.6 Track fragment box
//  aligned(8) class TrackFragmentBox extends Box('traf') {
//  }
type TrackFragmentBox struct {
   Header BoxHeader
   Boxes  []Box
   Trun   TrackRunBox
   Senc   SampleEncryptionBox
}
