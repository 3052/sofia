# mp4

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

split:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 6M -movflags empty_moov frag.mp4
~~~

join:

~~~
ffmpeg -i frag.mp4 -c copy -movflags default_base_moof out.mp4
~~~

everything reports as 8m1s. resulting file has no `trun` boxes. other boxes:

~~~
[stsc] size=472 version=0 flags=000000
- entryCount: 38
- entry[1]: firstChunk=1 samplesPerChunk=327 sampleDescriptionID=1

[stsz] size=57684 version=0 flags=000000
- sampleCount: 14416

[stco] size=168 version=0 flags=000000
- entryCount: 38
- entry[1]: chunkOffset=48

[moov] size=161342
  [trak] size=161128
    [mdia] size=160992
      [minf] size=160881
        [stbl] size=160817
          [stsc] size=472 version=0 flags=000000
          [stsz] size=57684 version=0 flags=000000
          [stco] size=168 version=0 flags=000000
~~~

https://godocs.io/github.com/Eyevinn/mp4ff/mp4#File
