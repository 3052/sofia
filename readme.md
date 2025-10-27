# sofia

> And, in the dream, I knew that he was, going on ahead. He was fixin' to make
> a fire somewhere out there in all that dark and cold. And I knew that
> whenever I got there, he'd be there. And then I woke up.
>
> [No Country for Old Men](//youtube.com/watch?v=GH4IhjtaAUQ) (2007)

## features

1. Firefox playback
2. decrypt `mdat` using `senc`
3. multiple `moof` boxes
4. parse `sidx`
5. remove `edts`
6. remove `pssh`
7. remove `sinf`
8. rename `enca`
9. rename `encv`
10. get bandwidth from `traf`

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

## standard

ISO/IEC 14496-12:

<https://wikipedia.org/wiki/ISO_base_media_file_format>

ISO/IEC 14496-14:

<https://wikipedia.org/wiki/MP4_file_format>

ISO/IEC 23001-7:

<https://wikipedia.org/wiki/MPEG_Common_Encryption>
