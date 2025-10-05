package main

import (
   "41.neocities.org/sofia/file"
   "encoding/hex"
   "flag"
   "log"
   "os"
)

func write_file(name string, data []byte) error {
   log.Println("WriteFile", name)
   return os.WriteFile(name, data, os.ModePerm)
}

func (f *flag_set) do_initialization() ([]byte, error) {
   data, err := os.ReadFile(f.initialization)
   if err != nil {
      return nil, err
   }
   var fileVar file.File
   err = fileVar.Read(data)
   if err != nil {
      return nil, err
   }
   for _, pssh := range fileVar.Moov.Pssh {
      copy(pssh.BoxHeader.Type[:], "free") // Firefox
   }
   description := fileVar.Moov.Trak.Mdia.Minf.Stbl.Stsd
   if sinf, ok := description.Sinf(); ok {
      // Firefox
      copy(sinf.BoxHeader.Type[:], "free")
      if sample, ok := description.SampleEntry(); ok {
         // Firefox
         copy(sample.BoxHeader.Type[:], sinf.Frma.DataFormat[:])
      }
   }
   return fileVar.Append(nil)
}

func main() {
   var set flag_set
   flag.StringVar(&set.initialization, "i", "", "initialization")
   flag.StringVar(&set.key, "k", "", "key")
   flag.StringVar(&set.output, "o", "", "output")
   flag.StringVar(&set.segment, "s", "", "segment")
   flag.Parse()
   if set.output != "" {
      var (
         data []byte
         err  error
      )
      if set.initialization != "" {
         data, err = set.do_initialization()
         if err != nil {
            panic(err)
         }
      }
      if set.segment != "" {
         data, err = set.do_segment(data)
         if err != nil {
            panic(err)
         }
      }
      err = write_file(set.output, data)
      if err != nil {
         panic(err)
      }
   } else {
      flag.Usage()
   }
}

type flag_set struct {
   initialization string
   key            string
   output         string
   segment        string
}

func (f *flag_set) do_segment(data []byte) ([]byte, error) {
   data1, err := os.ReadFile(f.segment)
   if err != nil {
      return nil, err
   }
   var fileVar file.File
   err = fileVar.Read(data1)
   if err != nil {
      return nil, err
   }
   track := fileVar.Moof.Traf
   if senc := track.Senc; senc != nil {
      key, err := hex.DecodeString(f.key)
      if err != nil {
         return nil, err
      }
      for i, data1 := range fileVar.Mdat.Data(&track) {
         err := senc.Sample[i].Decrypt(data1, key)
         if err != nil {
            return nil, err
         }
      }
   }
   return fileVar.Append(data)
}
