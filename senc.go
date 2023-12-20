package sofia

// aligned(8) class SampleEncryptionBox extends FullBox(
//    'senc',
//    version,
//    flags
// ) {
//    unsigned int(32) sample_count;
//    {
//       if (version==0) {
//          unsigned int(Per_Sample_IV_Size*8) InitializationVector;
//          if (UseSubSampleEncryption) {
//             unsigned int(16) subsample_count;
//             {
//                unsigned int(16) BytesOfClearData;
//                unsigned int(32) BytesOfProtectedData;
//             } [subsample_count ]
//          }
//       } else if (version==1 && isProtected) {
//          unsigned int(16) multi_IV_count;
//          for (i=1; i <= multi _IV_count; i++) {
//             unsigned int(16) multi_subindex_IV;
//             unsigned int(Per_Sample_IV_Size*8) IV;
//          }
//          unsigned int(32) subsample_count;
//          {
//             unsigned int(16) multi_subindex;
//             unsigned int(16) BytesOfClearData;
//             unsigned int(32) BytesOfProtectedData;
//          } [subsample_count]
//       } else if (version==2 && isProtected) {
//          unsigned int(Per_Sample_IV_Size*8) InitializationVector;
//          if (UseSubSampleEncryption) {
//             unsigned int(16) subsample_count;
//             {
//                unsigned int(16) BytesOfClearData;
//                unsigned int(32) BytesOfProtectedData;
//             } [subsample_count ]
//          }
//       }
//    }[ sample_count ]
// }
type SampleEncryptionBox struct{}
