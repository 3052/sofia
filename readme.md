# sofia

> And, in the dream, I knew that he was, going on ahead. He was fixin' to make
> a fire somewhere out there in all that dark and cold. And I knew that
> whenever I got there, he'd be there. And then I woke up.
>
> [No Country for Old Men](//youtube.com/watch?v=GH4IhjtaAUQ) (2007)

ISOBMFF

library for reading and writing MP4

## prior art

1. https://github.com/mozilla/mp4parse-rust/issues/415
2. https://github.com/Eyevinn/mp4ff/issues/311
3. https://github.com/alfg/mp4-rust/issues/132
4. https://github.com/yapingcat/gomedia/issues/115
5. https://github.com/alfg/mp4/issues/27
6. https://github.com/abema/go-mp4/issues/13
7. https://github.com/garden4hu/fmp4parser-go/issues/4
8. https://github.com/eswarantg/mp4box/issues/3
9. https://github.com/miquels/mp4/issues/2

> the normal place to put the encryption information in the segments is in a
> `senc` box, and this is not the case in this file, but they seem to be placed
> in a [uuid] box instead. This is allowed, but not supported by mp4ff library
> at the moment. In principle the data can be put in any place given by the
> offset in the `saio` box

and:

> If the Override TrackEncryptionBox parameters flag is set, then the
> SampleEncryptionBox specifies the `AlgorithmID`, `IV_size`, and `KID`
> parameters. If not present, then the default values from the
> TrackEncryptionBox SHOULD be used for this fragment.
