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

## progress

- [x] enca, provides sinf
- [x] encv, provides sinf
- [x] frma, provides original format
- [x] mdat, provides media data
- [x] mdia, provides minf
- [x] minf, provides stbl
- [x] moof, provides traf
- [x] moov, provides trak
- [x] pssh, provides pssh data
- [x] schi, provides tenc
- [x] senc, provides initialization vector
- [x] sidx, provides segment indexes
- [x] sinf, provides frma
- [x] stbl, provides stsd
- [x] stsd, provides enca encv
- [x] tenc, provides default key ID
- [x] tfhd, provides default sample size
- [x] traf, provides senc
- [x] trak, provides mdia
- [x] trun, provides sample sizes
