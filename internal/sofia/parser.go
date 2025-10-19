// File: parser.go
package mp4parser

import "fmt"

type Parser struct {
   data   []byte
   offset int
}

func NewParser(data []byte) *Parser { return &Parser{data: data} }
func (p *Parser) HasMore() bool     { return p.offset < len(p.data) }
func (p *Parser) ParseNextBox() (*Box, error) {
   if !p.HasMore() {
      return nil, nil
   }
   header, headerEndOffset, err := ParseBoxHeader(p.data, p.offset)
   if err != nil {
      return nil, err
   }
   resultBox := &Box{Header: header}
   contentEndOffset := p.offset + int(header.Size)
   if contentEndOffset > len(p.data) {
      return nil, ErrUnexpectedEOF
   }
   content := p.data[headerEndOffset:contentEndOffset]
   switch header.Type {
   case "moov":
      resultBox.Moov, err = ParseMoovBox(content)
   case "moof":
      resultBox.Moof, err = ParseMoofBox(content)
   case "mdat":
      resultBox.Mdat, err = ParseMdatBox(content)
   case "sidx":
      resultBox.Sidx, err = ParseSidxBox(content)
   default:
      resultBox.Raw, err = ParseRawBox(header.Type, content)
   }
   if err != nil {
      return nil, fmt.Errorf("failed to parse '%s' box at offset %d: %w", header.Type, p.offset, err)
   }
   p.offset = contentEndOffset
   return resultBox, nil
}
