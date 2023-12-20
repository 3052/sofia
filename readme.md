# sofia

ISOBMFF

- https://github.com/Eyevinn/mp4ff
- https://github.com/abema/go-mp4
- https://github.com/yapingcat/gomedia

## amc

init:

~~~
[moov] size=1950
  [trak] size=578
    [mdia] size=478
      [minf] size=385
        [stbl] size=321
          [stsd] size=245 version=0 flags=000000
            [encv] size=229
             - width: 1152
             - height: 648
             - compressorName: ""
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: bc791d3b-444f-4aca-83de-23f37aea4f78
~~~

segment:

~~~
[moof] size=6665
  [traf] size=6641
    [trun] size=1752 version=0 flags=000b05
     - sampleCount: 144
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
    [trun] size=1464 version=0 flags=000b05
     - sampleCount: 120
    [uuid] size=1952
     - uuid: a2394f52-5a9b-4f14-a244-6c427c648df4
     - subType: unknown
~~~

## nbc

init:

~~~
[moov] size=1819
  [trak] size=582
    [mdia] size=482
      [minf] size=389
        [stbl] size=325
          [stsd] size=249 version=0 flags=000000
            [encv] size=233
             - width: 1280
             - height: 720
             - compressorName: ""
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 0a95c346-cec8-4e49-9679-59580ab7789b
~~~

segment:

~~~
[moof] size=1889
  [traf] size=1865
    [trun] size=744 version=1 flags=000b05
     - sampleCount: 60
    [senc] size=976 version=0 flags=000002
     - sampleCount: 60
     - perSampleIVSize: 8
~~~

## paramount

init:

~~~
[moov] size=1866
  [trak] size=598
    [mdia] size=462
      [minf] size=377
        [stbl] size=313
          [stsd] size=237 version=0 flags=000000
            [encv] size=221
             - width: 1280
             - height: 720
             - compressorName: "AVC Coding"
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 2ae3928e-7686-4505-aa84-99db218b0288
~~~

segment:

~~~
[moof] size=4177
  [traf] size=4153
    [trun] size=1748 version=0 flags=000e01
     - sampleCount: 144
    [senc] size=2320 version=0 flags=000002
     - sampleCount: 144
     - perSampleIVSize: 8
~~~

## roku

init:

~~~
[moov] size=1489
  [trak] size=605
    [mdia] size=505
      [minf] size=405
        [stbl] size=341
          [stsd] size=265 version=0 flags=000000
            [encv] size=249
             - width: 1280
             - height: 720
             - compressorName: "Elemental H.264"
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: a965fe62-4f17-7ae2-3a0d-cd0097a813e9
~~~

segment:

~~~
[moof] size=2574
  [traf] size=1849
    [trun] size=600 version=1 flags=000b05
     - sampleCount: 48
    [senc] size=1072 version=0 flags=000002
     - sampleCount: 48
     - perSampleIVSize: 8
~~~
