# mp4

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

create:

~~~
ffmpeg -i in.mp4 -c copy -movflags default_base_moof out.mp4
~~~

everything reports as 8m1s. resulting file has no `trun` boxes. other boxes:

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
