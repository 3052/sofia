# paramount

~~~
paramount -b KtQCLnOCgoQiTRF_uGCbqesqmz7SjwRm -d -v debug -vb 147465
~~~

## init

~~~
[moov] size=1866
  [trak] size=598
    [mdia] size=462
      [minf] size=377
        [stbl] size=313
          [stsd] size=237 version=0 flags=000000
            [encv] size=221
             - width: 416
             - height: 234
             - compressorName: "AVC Coding"
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 2ae3928e-7686-4505-aa84-99db218b0288
~~~

## segment

~~~
[moof] size=4177
  [traf] size=4153
    [tfhd] size=24 version=0 flags=02000a
     - trackID: 1
     - defaultBaseIsMoof: true
     - sampleDescriptionIndex: 1
     - defaultSampleDuration: 1001
    [trun] size=1748 version=0 flags=000e01
     - sampleCount: 144
     - DataOffset: 4185
     - sample[1]: size=1060 flags=00000000 (isLeading=0 dependsOn=0 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=false degradationPriority=0) compositionTimeOffset=2002
     - sample[2]: size=267 flags=00010000 (isLeading=0 dependsOn=0 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=true degradationPriority=0) compositionTimeOffset=5005
    [senc] size=2320 version=0 flags=000002
     - sampleCount: 144
     - perSampleIVSize: 8
[mdat] size=59078
~~~
