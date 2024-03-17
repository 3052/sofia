package sofia

// ISO/IEC 23001-7
//  aligned(8) class ProtectionSystemSpecificHeaderBox extends FullBox(
//     'pssh', version, flags=0,
//  ) {
//     unsigned int(8)[16] SystemID;
//     if (version > 0) {
//        unsigned int(32) KID_count;
//        {
//           unsigned int(8)[16] KID;
//        } [KID_count];
//     }
//     unsigned int(32) DataSize;
//     unsigned int(8)[DataSize] Data;
//  }
type ProtectionSystemSpecificHeader struct {
   BoxHeader  BoxHeader
   FullBoxHeader FullBoxHeader
   SystemId [16]uint8
   DataSize uint32
   Data []uint8
}
