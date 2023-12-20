# nbc

~~~
nbc -b 9000283421 -v debug -m 359387
~~~

## init

~~~
[moov] size=1819
  [trak] size=582
    [mdia] size=482
      [minf] size=389
        [stbl] size=325
          [stsd] size=249 version=0 flags=000000
            [encv] size=233
             - width: 512
             - height: 288
             - compressorName: ""
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 0a95c346-cec8-4e49-9679-59580ab7789b
~~~

## segment

~~~
[moof] size=1889
  [traf] size=1865
    [trun] size=744 version=1 flags=000b05
     - sampleCount: 60
     - DataOffset: 1897
     - firstSampleFlags: 02000000 (isLeading=0 dependsOn=2 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=false degradationPriority=0)
    [senc] size=976 version=0 flags=000002
     - sampleCount: 60
     - perSampleIVSize: 8
~~~
