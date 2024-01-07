# fmp4

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

if you do this:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 6M frag.mp4
~~~

Windows reports as 1m10s, FFmpeg and MPC-HC reports as 8m1s. If you remove the
`trun` boxes, then everything reports as 1m10s. you can correct Windows by
editing `mvhd`. other boxes:

~~~
[stsz] size=8500 version=0 flags=000000
- sampleCount: 2120

[stsc] size=88 version=0 flags=000000
- entryCount: 6
- entry[1]: firstChunk=1 samplesPerChunk=327 sampleDescriptionID=1

[stco] size=40 version=0 flags=000000
- entryCount: 6
- entry[1]: chunkOffset=24434
~~~
