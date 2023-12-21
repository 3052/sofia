# roku

~~~
roku -b 1ad93a236d88595b86d312eb04e3646c -v debug -vb 0
~~~

## init

~~~
[moov] size=1394
  [trak] size=510
    [mdia] size=410
      [minf] size=310
        [stbl] size=250
          [stsd] size=174 version=0 flags=000000
            [enca] size=158
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: a965fe62-4f17-7ae2-3a0d-cd0097a813e9
~~~

## segment

~~~
[moof] size=1958
  [traf] size=1233
    [tfhd] size=20 version=0 flags=020020
     - trackID: 1
     - defaultBaseIsMoof: true
     - defaultSampleFlags: 0aa00000 (isLeading=2 dependsOn=2 isDependedOn=2 hasRedundancy=2 padding=0 isNonSync=false degradationPriority=0)
    [trun] size=632 version=1 flags=000b01
     - sampleCount: 51
     - DataOffset: 1966
     - sample[1]: dur=2048 size=682 compositionTimeOffset=0
     - sample[2]: dur=2048 size=683 compositionTimeOffset=0
    [senc] size=424 version=0 flags=000000
     - sampleCount: 51
     - perSampleIVSize: 8
~~~
