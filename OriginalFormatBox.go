package sofia

// Container: ProtectionSchemeInfoBox
//  aligned(8) class OriginalFormatBox(codingname) extends Box('frma') {
//     unsigned int(32) data_format = codingname;
//     // format of decrypted, encoded data (in case of protection)
//     // or un-transformed sample entry (in case of restriction
//     // and complete track information)
//  }
type OriginalFormatBox struct {
   Header BoxHeader
   Data_Format uint32
}
