# hulu

~~~
hulu -a hulu.com/watch/023c49bf-6a99-4c67-851c-4c9e7609cc1d -b 437551
~~~

## init

~~~
[moov] size=1502
  [trak] size=584
    [mdia] size=472
      [minf] size=375
        [stbl] size=311
          [stsd] size=235 version=0 flags=000000
            [encv] size=219
             - width: 512
             - height: 288
             - compressorName: ""
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 21b82dc2-ebb2-4d5a-a9f8-631f04726650
~~~

## segment

~~~
[moof] size=3641
  [traf] size=3617
    [tfhd] size=20 version=0 flags=020020
     - trackID: 1
     - defaultBaseIsMoof: true
     - defaultSampleFlags: 01010000 (isLeading=0 dependsOn=1 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=true degradationPriority=0)
    [trun] size=1464 version=0 flags=000b05
     - sampleCount: 120
     - DataOffset: 3649
     - firstSampleFlags: 02000000 (isLeading=0 dependsOn=2 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=false degradationPriority=0)
     - sample[1]: dur=417083 size=79 compositionTimeOffset=834167
     - sample[2]: dur=417084 size=20 compositionTimeOffset=2502500
    [uuid] size=1952
     - uuid: a2394f52-5a9b-4f14-a244-6c427c648df4
     - subType: unknown
[mdat] size=186478
~~~
