package ttml

import (
   "154.pages.dev/sofia"
   "errors"
   "io"
   "encoding/xml"
   "strings"
)

const web_vtt = "WEBVTT"

func (p paragraph) String() string {
   var b strings.Builder
   b.WriteString("\n\n")
   b.WriteString(p.Begin)
   b.WriteString(" --> ")
   b.WriteString(p.End)
   b.WriteByte('\n')
   b.WriteString(p.Text)
   return b.String()
}

type paragraph struct {
   Begin  string `xml:"begin,attr"`
   End    string `xml:"end,attr"`
   Text   string `xml:",chardata"`
}

type timed_text struct {
   Body struct {
      Div  struct {
         P []paragraph `xml:"p"`
      } `xml:"div"`
   } `xml:"body"`
}

func (t *timed_text) decode(r io.Reader) error {
   var file sofia.File
   err := file.Decode(r)
   if err != nil {
      return err
   }
   if len(file.MediaData.Data) != 1 {
      return errors.New("sofia.File.MediaData.Data")
   }
   return xml.Unmarshal(file.MediaData.Data[0], t)
}
