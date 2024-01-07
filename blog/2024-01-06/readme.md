# jan 3 2024

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

if you do this:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 6M frag.mp4
~~~

Windows reports as 1m10s, FFmpeg and MPC-HC reports as 8m1s. If you remove the
`trun` boxes, then everything reports as 1m10s. if you do this:

~~~
ffmpeg -i frag.mp4 -c copy -movflags default_base_moof out.mp4
~~~

everything reports as 8m1s. resulting file has no `trun` boxes. result with
`-frag_size`:

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

result without `-frag_size`:

~~~
[stsz] size=57684 version=0 flags=000000
- sampleCount: 14416

[stsc] size=472 version=0 flags=000000
- entryCount: 38
- entry[1]: firstChunk=1 samplesPerChunk=327 sampleDescriptionID=1

[stco] size=168 version=0 flags=000000
- entryCount: 38
- entry[1]: chunkOffset=48
~~~
