# amc

~~~
amc `
-a amcplus.com/shows/orphan-black/episodes/season-1-instinct--1011152 `
-h 0 `
-v debug
~~~

## init

~~~
[moov] size=1875
  [trak] size=503
    [mdia] size=403
      [minf] size=310
        [stbl] size=250
          [stsd] size=174 version=0 flags=000000
            [enca] size=158
              [sinf] size=80
                [schi] size=40
                  [tenc] size=32 version=0 flags=000000
                   - defaultIsProtected: 1
                   - defaultPerSampleIVSize: 8
                   - defaultKID: 5e7d369b-9eca-4426-a43e-15a76f09dd7e
~~~

## segment

~~~
[moof] size=4641
  [traf] size=4617
    [trun] size=2252 version=0 flags=000301
     - sampleCount: 279
     - DataOffset: 6881
     - sample[1]: dur=1024 size=17
     - sample[2]: dur=1024 size=20
    [senc] size=16 version=0 flags=000000
     - sampleCount: 279
     - perSampleIVSize: 0
    [uuid] size=2264
     - uuid: a2394f52-5a9b-4f14-a244-6c427c648df4
     - subType: unknown
~~~
