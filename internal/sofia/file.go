package sofia

import (
   "io"
   "sofia/box"
   "sofia/mdat"
)

// ISO/IEC 14496-12
type File struct {
   Boxes         []box.Box
   MediaData     *mdat.Box
}

func (f *File) Read(r io.Reader) error {
   for {
      var head box.Header
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "mdat":
            f.MediaData = &mdat.Box{}
            f.MediaData.Box.BoxHeader = head
            err := f.MediaData.Read(r)
            if err != nil {
               return err
            }
         case "free", // Mubi
         "ftyp", // Roku
         "styp": // Roku
            object := box.Box{BoxHeader: head}
            err := object.Read(r)
            if err != nil {
               return err
            }
            f.Boxes = append(f.Boxes, object)
         default:
            var container box.Type
            copy(container[:], "File")
            return box.Error{container, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}
