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

SegmentTemplate:

https://play.google.com/store/apps/details?id=com.roku.remote

example:

https://therokuchannel.roku.com/watch/597a64a4a25c5bf6af4a8c7053049a6f

~~~
N_m3u8DL-RE --skip-merge `
'https://vod-playlist.sr.roku.com/1.mpd?origin=https%3A%2F%2Fvod.delivery.roku.com%2F4eb6f71c-0374-4403-8e73-83334c1ca62b%2Findex-1749161125402.mpd%3Faws.manifestfilter%3Daudio_language%3Aen%2Ceng'

Vid *CENC 1920x1080 | 3759 Kbps | 7 | 23.976 | avc1.640028 | 994 Segments | ~01h39m24s
~~~

## decrypt

~~~
18:03:28 key ID 28339ad78f734520da24e6e0573d392e
18:03:28 DASH content ID 2a
18:03:28 MP4 content ID 2a
18:03:28 key 13d7c7cf295444944b627ef0ad2c1b3c
~~~
