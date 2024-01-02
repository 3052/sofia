package stream

import (
   "154.pages.dev/sofia"
   "bytes"
   "io"
)

func decode_sidx(data []byte, sidx, moof uint32) ([][2]uint32, error) {
   var f sofia.File
   if err := f.Decode(bytes.NewReader(data[sidx:moof])); err != nil {
      return nil, err
   }
   return f.SegmentIndex.ByteRanges(moof), nil
}

func encode_segment(dst io.Writer, src io.Reader, key []byte) error {
   var f sofia.File
   if err := f.Decode(src); err != nil {
      return err
   }
   for i, data := range f.MediaData.Data {
      sample := f.MovieFragment.TrackFragment.SampleEncryption.Samples[i]
      err := sample.Decrypt_CENC(data, key)
      if err != nil {
         return err
      }
   }
   return f.Encode(dst)
}

func encode_init(dst io.Writer, src io.Reader) error {
   var f sofia.File
   if err := f.Decode(src); err != nil {
      return err
   }
   for _, b := range f.Movie.Boxes {
      if b.Header.BoxType() == "pssh" {
         copy(b.Header.Type[:], "free") // Firefox
      }
   }
   sd := &f.Movie.Track.Media.MediaInformation.SampleTable.SampleDescription
   if as := sd.AudioSample; as != nil {
      copy(as.ProtectionScheme.Header.Type[:], "free") // Firefox
      copy(
         as.Entry.Header.Type[:],
         as.ProtectionScheme.OriginalFormat.DataFormat[:],
      ) // Firefox
   }
   if vs := sd.VisualSample; vs != nil {
      copy(vs.ProtectionScheme.Header.Type[:], "free") // Firefox
      copy(
         vs.Entry.Header.Type[:],
         vs.ProtectionScheme.OriginalFormat.DataFormat[:],
      ) // Firefox
   }
   return f.Encode(dst)
}
