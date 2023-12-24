package sofia

import (
   "encoding/binary"
   "fmt"
   "io"
)

// aligned(8) abstract class SampleEntry(unsigned int(32) format) extends Box(format) {
//    const unsigned int(8)[6] reserved = 0;
//    unsigned int(16) data_reference_index;
// }
type SampleEntry struct {
   Header  BoxHeader
   Reserved [6]uint8
   Data_Reference_Index uint16
}

// class VisualSampleEntry(codingname) extends SampleEntry(codingname) {
//    unsigned int(16) pre_defined = 0;
//    const unsigned int(16) reserved = 0;
//    unsigned int(32)[3] pre_defined = 0;
//    unsigned int(16) width;
//    unsigned int(16) height;
//    template unsigned int(32) horizresolution = 0x00480000; // 72 dpi
//    template unsigned int(32) vertresolution = 0x00480000; // 72 dpi
//    const unsigned int(32) reserved = 0;
//    template unsigned int(16) frame_count = 1;
//    uint(8)[32] compressorname;
//    template unsigned int(16) depth = 0x0018;
//    int(16) pre_defined = -1;
//    // other boxes from derived specifications
//    CleanApertureBox clap; // optional
//    PixelAspectRatioBox pasp; // optional
// }
type VisualSampleEntry struct {
   Entry SampleEntry
   Pre_Defined uint16
   Reserved uint16
   Pre_Defined [3]uint32
   Width uint16
   Height uint16
   HorizResolution uint32
   VertResolution uint32
   Reserved uint32
   Frame_Count uint16
   CompressorName [32]uint8
   Depth uint16
   Pre_Defined int16
}
