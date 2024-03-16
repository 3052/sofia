package text

import (
   "154.pages.dev/sofia"
   "errors"
   "html"
   "io"
   "encoding/xml"
   "strings"
)

const WebVtt = "WEBVTT"

type Markup struct {
   Body struct {
      Div  struct {
         P []Paragraph `xml:"p"`
      } `xml:"div"`
   } `xml:"body"`
}

func (m *Markup) Decode(r io.Reader) error {
   var file sofia.File
   err := file.Decode(r)
   if err != nil {
      return err
   }
   if len(file.MediaData.Data) != 1 {
      return errors.New("sofia.MediaDataBox.Data")
   }
   return xml.Unmarshal(file.MediaData.Data[0], m)
}

type Paragraph struct {
   Begin  string `xml:"begin,attr"`
   End    string `xml:"end,attr"`
   Text   string `xml:",innerxml"`
}

func (p Paragraph) String() string {
   p.Text = html.UnescapeString(p.Text)
   p.Text = strings.ReplaceAll(p.Text, "<br />", "\n")
   var b strings.Builder
   b.WriteByte('\n')
   b.WriteString(p.Begin)
   b.WriteString(" --> ")
   b.WriteString(p.End)
   b.WriteByte('\n')
   b.WriteString(p.Text)
   return b.String()
}
