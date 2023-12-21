# amc

~~~
amc `
-a amcplus.com/shows/orphan-black/episodes/season-1-instinct--1011152 `
-h 216 `
-v debug
~~~

## init

~~~
[moov] size=1948
  [trak] size=576
    [mdia] size=476
      [minf] size=383
        [stbl] size=319
          [stsd] size=243 version=0 flags=000000
            [encv] size=227
             - width: 384
             - height: 216
             - compressorName: ""
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: bc791d3b-444f-4aca-83de-23f37aea4f78
~~~

## segment

~~~
[moof] size=6665
  [traf] size=6641
    [tfhd] size=24 version=0 flags=020028
     - trackID: 1
     - defaultBaseIsMoof: true
     - defaultSampleDuration: 1001
     - defaultSampleFlags: 01010000 (isLeading=0 dependsOn=1 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=true degradationPriority=0)
    [trun] size=1752 version=0 flags=000b05
     - sampleCount: 144
     - DataOffset: 6673
     - firstSampleFlags: 02000000 (isLeading=0 dependsOn=2 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=false degradationPriority=0)
     - sample[1]: dur=1001 size=97 compositionTimeOffset=2002
     - sample[2]: dur=1001 size=2883 compositionTimeOffset=4004
    [senc] size=2320 version=0 flags=000002
     - sampleCount: 144
     - perSampleIVSize: 8
    [uuid] size=2336
     - uuid: a2394f52-5a9b-4f14-a244-6c427c648df4
     - subType: unknown
[mdat] size=179944
~~~
