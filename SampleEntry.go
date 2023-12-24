package sofia

// All SampleEntry boxes may contain “extra boxes” not explicitly defined in the
// box syntax of this or derived specifications. When present, such boxes shall
// follow all defined fields and should follow any defined contained boxes.
// Decoders shall presume a sample entry box could contain extra boxes and shall
// continue parsing as though they are present until the containing box length is
// exhausted.
//
// aligned(8) abstract class SampleEntry(unsigned int(32) format) extends Box(format) {
//    const unsigned int(8)[6] reserved = 0;
//    unsigned int(16) data_reference_index;
// }
type SampleEntry struct {
   Header BoxHeader
   Reserved [6]uint8
   Data_Reference_Index uint16
   Boxes []Box
}
