# mp4

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

split:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 6M -movflags empty_moov frag.mp4

1
ffmpeg -i in.mp4 -c copy -frag_size 39M -y -movflags empty_moov frag.mp4
~~~

join:

~~~
ffmpeg -i frag.mp4 -c copy -movflags default_base_moof out.mp4
~~~

everything reports as 8m1s. can we kill `edts`? yes. can we kill `stss`? yes. can
we kill `ctts`? yes.

https://godocs.io/github.com/Eyevinn/mp4ff/mp4#File

how do we fill `stsz`? input looks like this:

~~~
[moof] size=25548
  [traf] size=25524
    [trun] size=25460 version=0 flags=000e01
     - sampleCount: 2120
     - DataOffset: 25556
     - sample[1]: size=520 flags=02000000 (isLeading=0 dependsOn=2 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=false degradationPriority=0) compositionTimeOffset=1001
~~~

output:

~~~
[moov] size=161342
  [trak] size=161128
    [mdia] size=160992
      [minf] size=160881
        [stbl] size=160817
          [stsz] size=57684 version=0 flags=000000
           - sampleCount: 14416
           - sample[1] size=520
~~~

first:

~~~
[stsz] size=57684 version=0 flags=000000
- sampleCount: 14416
~~~

second:

~~~
[stco] size=168 version=0 flags=000000
- entryCount: 38
- entry[1]: chunkOffset=48
~~~

third:

~~~
[stsc] size=472 version=0 flags=000000
- entryCount: 38
- entry[1]: firstChunk=1 samplesPerChunk=327 sampleDescriptionID=1
~~~
