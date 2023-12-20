# sofia

ISOBMFF

- https://github.com/Eyevinn/mp4ff
- https://github.com/abema/go-mp4
- https://github.com/yapingcat/gomedia

## amc

~~~
[moov] size=1937
  [trak] size=565
    [mdia] size=465
      [minf] size=372
        [stbl] size=308
          [stsd] size=232 version=0 flags=000000
            [encv] size=216
             - width: 384
             - height: 216
             - compressorName: ""
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 5c222a3e-2cfb-4b86-9773-ea680f1f3363
[moof] size=6665
  [traf] size=6641
    [senc] size=2320 version=0 flags=000002
     - sampleCount: 144
     - perSampleIVSize: 8
    [uuid] size=2336
     - uuid: a2394f52-5a9b-4f14-a244-6c427c648df4
     - subType: unknown
~~~

## hulu

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
[moof] size=3641
  [traf] size=3617
    [uuid] size=1952
     - uuid: a2394f52-5a9b-4f14-a244-6c427c648df4
     - subType: unknown
~~~

## nbc

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
                   - defaultKID: 0f8f0b8a-ff43-4541-8a1e-72162017884e
[moof] size=1889
  [traf] size=1865
    [senc] size=976 version=0 flags=000002
     - sampleCount: 60
     - perSampleIVSize: 8
~~~

## paramount

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
                   - defaultKID: bf9eeb01-706e-4067-ac06-3e15c3ba38d0
[moof] size=4177
  [traf] size=4153
    [senc] size=2320 version=0 flags=000002
     - sampleCount: 144
     - perSampleIVSize: 8
~~~

## roku

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
                   - defaultKID: bdfa4d6c-db39-702e-5b68-1f90617f9a7e
[moof] size=2574
  [traf] size=1849
    [senc] size=1072 version=0 flags=000002
     - sampleCount: 48
     - perSampleIVSize: 8
~~~
