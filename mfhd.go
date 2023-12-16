package sofia

// aligned(8) class FullBox(
//    unsigned int(32) boxtype,
//    unsigned int(8) v, bit(24) f,
//    optional unsigned int(8)[16] extended_type
// ) extends Box(boxtype, extended_type) {
//    FullBoxHeader(v, f);
//    // the remaining bytes are the FullBoxPayload
// }
type FullBox struct{}

// aligned(8) class MovieFragmentHeaderBox extends FullBox('mfhd', 0, 0) {
//    unsigned int(32) sequence_number;
// }
type MovieFragmentHeader struct{}
