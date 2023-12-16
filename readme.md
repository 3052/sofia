# sofia

ISOBMFF

Roku:

~~~
[moof] size=2574
  [traf] size=1849
    [senc] size=1072 version=0 flags=000002
     - sampleCount: 48
     - perSampleIVSize: 8
~~~

## Eyevinn/mp4ff

this works:

~~~
mp4ff-info index_video_5_0_1.mp4
~~~

https://github.com/Eyevinn/mp4ff

## abema/go-mp4

this works:

~~~
mp4tool dump index_video_5_0_1.mp4
~~~

https://github.com/abema/go-mp4

## yapingcat/gomedia

these all fail:

~~~
go run example_demux_fmp4.go index_video_5_0_1.mp4
go run example_demux_mp4.go -mp4file index_video_5_0_1.mp4
go run example_demux_mp4_memeory_io.go -mp4file index_video_5_0_1.mp4
~~~

https://github.com/yapingcat/gomedia
