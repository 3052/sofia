package sofia

// 8.16.3 Segment index box
//  aligned(8) class SegmentIndexBox extends FullBox('sidx', version, 0) {
//     unsigned int(32) reference_ID;
//     unsigned int(32) timescale;
//     if (version==0) {
//        unsigned int(32) earliest_presentation_time;
//        unsigned int(32) first_offset;
//     } else {
//        unsigned int(64) earliest_presentation_time;
//        unsigned int(64) first_offset;
//     }
//     unsigned int(16) reserved = 0;
//     unsigned int(16) reference_count;
//     for(i=1; i <= reference_count; i++) {
//        bit (1) reference_type;
//        unsigned int(31) referenced_size;
//        unsigned int(32) subsegment_duration;
//        bit(1) starts_with_SAP;
//        unsigned int(3) SAP_type;
//        unsigned int(28) SAP_delta_time;
//     }
//  }
type SegmentIndexBox struct {
   BoxHeader BoxHeader
   FullBoxHeader FullBoxHeader
   Reference_ID uint32
   Timescale uint32
   Earliest_Presentation_Time []byte
   First_Offset []byte
   Reserved uint16
   Reference_Count uint16
   References []Reference
}

type Reference struct {
   Reference_Type bool
   Referenced_Size uint32
   Subsegment_Duration uint32
   Starts_With_SAP bool
   SAP_type uint8
   SAP_delta_time uint32
}
