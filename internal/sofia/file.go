package sofia

import "io"

// ISO/IEC 14496-12
type File struct {
   Boxes         []Box
   MediaData     *MediaData
   Movie         *Movie
   MovieFragment *MovieFragment
   SegmentIndex  *SegmentIndex
}

func (f *File) Read(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         _, size := head.get_size()
         switch head.Type.String() {
         case "mdat":
            f.MediaData = &MediaData{}
            f.MediaData.Box.BoxHeader = head
            err := f.MediaData.read(r)
            if err != nil {
               return err
            }
         case "free", // Mubi
         "ftyp", // Roku
         "styp": // Roku
            object := Box{BoxHeader: head}
            err := object.read(r)
            if err != nil {
               return err
            }
            f.Boxes = append(f.Boxes, object)
         default:
            var container Type
            copy(container[:], "File")
            return box_error{container, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}
