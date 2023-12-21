# paramount

~~~
paramount -b KtQCLnOCgoQiTRF_uGCbqesqmz7SjwRm -d -v debug -vb 0
~~~

## init

~~~
[moov] size=1786
  [trak] size=518
    [mdia] size=418
      [minf] size=333
        [stbl] size=273
          [stsd] size=171 version=0 flags=000000
            [enca] size=155
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 2ae3928e-7686-4505-aa84-99db218b0288
~~~

## segment

~~~
[moof] size=1305
  [traf] size=1281
    [tfhd] size=28 version=0 flags=02002a
     - trackID: 1
     - defaultBaseIsMoof: true
     - sampleDescriptionIndex: 1
     - defaultSampleDuration: 1024
     - defaultSampleFlags: 00000000 (isLeading=0 dependsOn=0 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=false degradationPriority=0)
    [trun] size=1148 version=0 flags=000201
     - sampleCount: 282
     - DataOffset: 3569
     - sample[1]: size=335
     - sample[2]: size=334
    [senc] size=16 version=0 flags=000000
     - sampleCount: 282
     - perSampleIVSize: 0
[mdat] size=94632
~~~
