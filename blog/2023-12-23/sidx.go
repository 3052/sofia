package sidx

import (
   "bytes"
   "crypto/aes"
   "crypto/cipher"
   "errors"
   "fmt"
   "github.com/yapingcat/gomedia/go-mp4"
   "io"
   "net/http"
)

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func Decrypt_CENC(sample []byte, key []byte, subSample *mp4.SubSample) error {
   block, err := aes.NewCipher(key)
   if err != nil {
      return err
   }
   stream := cipher.NewCTR(block, subSample.IV[:])
   if len(subSample.Patterns) != 0 {
      var pos uint32 = 0
      for j := 0; j < len(subSample.Patterns); j++ {
         ss := subSample.Patterns[j]
         nrClear := uint32(ss.BytesClear)
         if nrClear > 0 {
            pos += nrClear
         }
         nrEnc := ss.BytesProtected
         if nrEnc > 0 {
            stream.XORKeyStream(sample[pos:pos+nrEnc], sample[pos:pos+nrEnc])
            pos += nrEnc
         }
      }
   } else {
      stream.XORKeyStream(sample, sample)
   }
   return nil
}

func get(url string, start, end uint32) ([]byte, error) {
   req, err := http.NewRequest("GET", url, nil)
   if err != nil {
      return nil, err
   }
   req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", start, end))
   fmt.Println(start, end)
   res, err := http.DefaultClient.Do(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusPartialContent {
      return nil, errors.New(res.Status)
   }
   return io.ReadAll(res.Body)
}

func byte_ranges(r io.Reader, start uint32) ([][]uint32, error) {
   sidx := mp4.SegmentIndexBox{
      Box: &mp4.FullBox{
         Box: &mp4.BasicBox{},
      },
   }
   if _, err := sidx.Box.Box.Decode(r); err != nil {
      return nil, err
   }
   if _, err := sidx.Decode(r); err != nil {
      return nil, err
   }
   var rs [][]uint32
   for _, e := range sidx.Entrys {
      r := []uint32{start, start + e.ReferencedSize - 1}
      rs = append(rs, r)
      start += e.ReferencedSize
   }
   return rs, nil
}

func mux(
   dst io.WriteSeeker,
   url string,
   start_sidx, start_segment uint32,
   key []byte,
) error {
   muxer, err := mp4.CreateMp4Muxer(dst)
   if err != nil {
      return err
   }
   vid := muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
   init, err := get(url, 0, start_sidx-1)
   if err != nil {
      return err
   }
   sidx, err := get(url, start_sidx, start_segment-1)
   if err != nil {
      return err
   }
   ranges, err := byte_ranges(bytes.NewReader(sidx), start_segment)
   if err != nil {
      return err
   }
   for _, r := range ranges {
      segment, err := get(url, r[0], r[1])
      if err != nil {
         return err
      }
      segment = append(init, segment...)
      demuxer := mp4.CreateMp4Demuxer(bytes.NewReader(segment))
      if _, err := demuxer.ReadHead(); err != nil {
         return err
      }
      demuxer.OnRawSample = func(_ mp4.MP4_CODEC_TYPE, sample []byte, subSample *mp4.SubSample) error {
         return Decrypt_CENC(sample, key, subSample)
      }
      for {
         pkg, err := demuxer.ReadPacket()
         if err == io.EOF {
            break
         } else if err != nil {
            return err
         }
         if err := muxer.Write(vid, pkg.Data, pkg.Pts, pkg.Dts); err != nil {
            return err
         }
      }
   }
   return muxer.WriteTrailer()
}
