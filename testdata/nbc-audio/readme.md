# nbc

~~~
nbc -b 9000283421 -v debug -m 0
~~~

## init

~~~
[moov] size=1737
  [trak] size=500
    [mdia] size=400
      [minf] size=307
        [stbl] size=247
          [stsd] size=171 version=0 flags=000000
            [enca] size=155
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 0a95c346-cec8-4e49-9679-59580ab7789b
~~~

## segment

~~~
[moof] size=893
  [traf] size=869
    [trun] size=772 version=0 flags=000301
     - sampleCount: 94
     - DataOffset: 1653
    [senc] size=16 version=0 flags=000000
     - sampleCount: 94
     - perSampleIVSize: 0
~~~
