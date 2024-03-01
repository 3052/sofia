package internal

import (
   "154.pages.dev/encoding/dash"
   "154.pages.dev/log"
   "154.pages.dev/sofia"
   "154.pages.dev/widevine"
   "encoding/hex"
   "errors"
   "fmt"
   "io"
   "log/slog"
   "net/http"
   "net/url"
   "os"
   "slices"
)

func (h HttpStream) segment_template(
   ext, initialization string, rep dash.Representation,
) error {
   key, err := h.key(rep)
   if err != nil {
      return err
   }
   slog.Debug("hex", "key", hex.EncodeToString(key))
   req, err := http.NewRequest("GET", initialization, nil)
   if err != nil {
      return err
   }
   req.URL = h.base.ResolveReference(req.URL)
   res, err := http.DefaultClient.Do(req)
   if err != nil {
      return err
   }
   defer res.Body.Close()
   file, err := os.Create("dec.m4f")
   if err != nil {
      return err
   }
   defer file.Close()
   if err := encode_init(file, res.Body); err != nil {
      return err
   }
   media := rep.Media()
   var meter log.ProgressMeter
   meter.Set(len(media))
   log.TransportDebug()
   defer log.TransportInfo()
   for _, ref := range media {
      // with DASH, initialization and media URLs are relative to the MPD URL
      req.URL, err = h.base.Parse(ref)
      if err != nil {
         return err
      }
      err := func() error {
         res, err := http.DefaultClient.Do(req)
         if err != nil {
            return err
         }
         defer res.Body.Close()
         if res.StatusCode != http.StatusOK {
            return errors.New(res.Status)
         }
         return encode_segment(file, meter.Reader(res), key)
      }()
      if err != nil {
         return err
      }
   }
   return nil
}

func (h HttpStream) key(rep dash.Representation) ([]byte, error) {
   var protect widevine.PSSH
   data, err := rep.PSSH()
   if err != nil {
      key_id, err := rep.Default_KID()
      if err != nil {
         return nil, err
      }
      protect.Key_ID = key_id
   } else {
      err := protect.New(data)
      if err != nil {
         return nil, err
      }
   }
   private_key, err := os.ReadFile(h.Private_Key)
   if err != nil {
      return nil, err
   }
   client_id, err := os.ReadFile(h.Client_ID)
   if err != nil {
      return nil, err
   }
   module, err := protect.CDM(private_key, client_id)
   if err != nil {
      return nil, err
   }
   license, err := module.License(h.Poster)
   if err != nil {
      return nil, err
   }
   key, ok := module.Key(license)
   if !ok {
      return nil, errors.New("widevine.CDM.Key")
   }
   return key, nil
}

func encode_init(dst io.Writer, src io.Reader) error {
   var f sofia.File
   if err := f.Decode(src); err != nil {
      return err
   }
   for _, b := range f.Movie.Boxes {
      if b.BoxHeader.BoxType() == "pssh" {
         copy(b.BoxHeader.Type[:], "free") // Firefox
      }
   }
   sd := &f.Movie.Track.Media.MediaInformation.SampleTable.SampleDescription
   if as := sd.AudioSample; as != nil {
      copy(as.ProtectionScheme.BoxHeader.Type[:], "free") // Firefox
      copy(
         as.Entry.BoxHeader.Type[:],
         as.ProtectionScheme.OriginalFormat.DataFormat[:],
      ) // Firefox
   }
   if vs := sd.VisualSample; vs != nil {
      copy(vs.ProtectionScheme.BoxHeader.Type[:], "free") // Firefox
      copy(
         vs.Entry.BoxHeader.Type[:],
         vs.ProtectionScheme.OriginalFormat.DataFormat[:],
      ) // Firefox
   }
   return f.Encode(dst)
}

func encode_segment(dst io.Writer, src io.Reader, key []byte) error {
   var f sofia.File
   if err := f.Decode(src); err != nil {
      return err
   }
   for i, data := range f.MediaData.Data {
      sample := f.MovieFragment.TrackFragment.SampleEncryption.Samples[i]
      err := sample.DecryptCenc(data, key)
      if err != nil {
         return err
      }
   }
   return f.Encode(dst)
}

// wikipedia.org/wiki/Dynamic_Adaptive_Streaming_over_HTTP
type HttpStream struct {
   Client_ID string
   Poster widevine.Poster
   Private_Key string
   base *url.URL
}
