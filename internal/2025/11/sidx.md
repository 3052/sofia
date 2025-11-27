# sidx

Go language, given the attached package, if possible I need a method on
`SidxBox` where you give it `ReferencedSize` input, and it adds a
`SidxReference` to the slice, and updates any other fields as needed. also if
possible I would like to remove RawData from `SidxBox` and do a proper Encode.
please make sure the `SidxBox` Parse/Encode is still a proper round trip with
the update

https://github.com/Eyevinn/mp4ff/issues/311

## 1 download

SegmentBase:

https://play.google.com/store/apps/details?id=com.tubitv

SegmentBase:

https://play.google.com/store/apps/details?id=com.wbd.stream

SegmentBase:

https://play.google.com/store/apps/details?id=com.cbs.app

SegmentTemplate, but each segment has its own `sidx` so might be bad test:

https://play.google.com/store/apps/details?id=com.roku.remote

SegmentTemplate, but each segment has its own `sidx` so might be bad test:

https://play.google.com/store/apps/details?id=tv.pluto.android

https://play.google.com/store/apps/details?id=com.plexapp.android

---------------------------------------------------------------------------------

example:

https://pluto.tv/on-demand/movies/6495eff09263a40013cf63a5

~~~
N_m3u8DL-RE --skip-merge `
http://silo-hybrik.pluto.tv.s3.amazonaws.com/735_Paramount_Pictures_LF/clip/6495efee9263a40013cf638d_Jack_Reacher/1080pDRM/20241115_113001/dash/0-end/main.mpd

Vid *CENC 4586 Kbps | 8 | 30 | avc1.640028 | 1565 Segments | Main | ~02h10m24s
~~~

## decrypt

~~~
19:42:36 key ID 000000006737a1396ec1349107feb9f4
19:42:36 key d55db54e95d0e3df95981b37d2742cb7
~~~
