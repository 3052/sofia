# hulu

~~~
hulu -a hulu.com/watch/023c49bf-6a99-4c67-851c-4c9e7609cc1d -b 0
~~~

## init

~~~
[moov] size=1449
  [trak] size=531
    [mdia] size=419
      [minf] size=322
        [stbl] size=262
          [stsd] size=186 version=0 flags=000000
            [enca] size=170
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 21b82dc2-ebb2-4d5a-a9f8-631f04726650
~~~

## segment

~~~
[moof] size=2081
  [traf] size=2057
    [trun] size=984 version=0 flags=000305
     - sampleCount: 120
     - DataOffset: 2089
     - firstSampleFlags: 02000000 (isLeading=0 dependsOn=2 isDependedOn=0 hasRedundancy=0 padding=0 isNonSync=false degradationPriority=0)
     - sample[1]: dur=426667 size=342
     - sample[2]: dur=426666 size=341
    [uuid] size=992
     - uuid: a2394f52-5a9b-4f14-a244-6c427c648df4
     - subType: unknown
~~~
