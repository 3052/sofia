# roku

~~~
roku -b 1ad93a236d88595b86d312eb04e3646c -v debug -vb 148223
~~~

## init

~~~
[moov] size=1484
  [trak] size=600
    [mdia] size=500
      [minf] size=400
        [stbl] size=336
          [stsd] size=260 version=0 flags=000000
            [encv] size=244
             - width: 384
             - height: 216
             - compressorName: "Elemental H.264"
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: a965fe62-4f17-7ae2-3a0d-cd0097a813e9
~~~

## segment

~~~
[moof] size=2574
  [traf] size=1849
    [tfhd] size=20 version=0 flags=020020
     - trackID: 1
     - defaultBaseIsMoof: true
     - defaultSampleFlags: 00610000 (isLeading=0 dependsOn=0 isDependedOn=1 hasRedundancy=2 padding=0 isNonSync=true degradationPriority=0)
    [trun] size=600 version=1 flags=000b05
     - sampleCount: 48
     - DataOffset: 2582
     - firstSampleFlags: 02600000 (isLeading=0 dependsOn=2 isDependedOn=1 hasRedundancy=2 padding=0 isNonSync=false degradationPriority=0)
     - sample[1]: dur=1001 size=307 compositionTimeOffset=0
     - sample[2]: dur=1001 size=173 compositionTimeOffset=0
    [senc] size=1072 version=0 flags=000002
     - sampleCount: 48
     - perSampleIVSize: 8
~~~
