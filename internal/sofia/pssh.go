package mp4

// PsshBox represents the 'pssh' box (Protection System Specific Header).
type PsshBox struct {
   Header  BoxHeader
   RawData []byte // Stores the original box data for a perfect round trip
   // We can add parsed fields here later if needed, e.g., SystemID, Data
}

// ParsePssh parses the 'pssh' box from a byte slice.
func ParsePssh(data []byte) (PsshBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return PsshBox{}, err
   }
   var pssh PsshBox
   pssh.Header = header
   pssh.RawData = data[:header.Size] // Store the original data

   // Parsing of internal fields could be added here if necessary for other features.

   return pssh, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *PsshBox) Encode() []byte {
   return b.RawData
}
