# amc

~~~
amc `
-a amcplus.com/shows/orphan-black/episodes/season-1-instinct--1011152 `
-h 0 `
-v debug
~~~

## init

~~~
[moov] Size=1875
  [trak] Size=503
    [mdia] Size=403
      [minf] Size=310
        [stbl] Size=250
          [stsd] Size=174 Version=0 Flags=0x000000 EntryCount=1
            [enca] Size=158 DataReferenceIndex=1 EntryVersion=0 ChannelCount=2 SampleSize=16 PreDefined=0 SampleRate=48000
              [sinf] Size=80
                [schi] Size=40
                  [tenc] Size=32 Version=0 Flags=0x000000 Reserved=0 DefaultCryptByteBlock=0 DefaultSkipByteBlock=0 DefaultIsProtected=1 DefaultPerSampleIVSize=8 DefaultKID=5e7d369b-9eca-4426-a43e-15a76f09dd7e
~~~

## segment

~~~
[moof] Size=6873
  [traf] Size=6849
    [tfhd] Size=20 Version=0 Flags=0x020008 TrackID=1 DefaultSampleDuration=1024
    [trun] Size=2252 Version=0 Flags=0x000301 SampleCount=279 DataOffset=6881 
    [senc] (unsupported box type) Size=2248 Data=[...] (use "-full senc" to show all)
    [uuid] (unsupported box type) Size=2264 Data=[...] (use "-full uuid" to show all)
[mdat] Size=91874 Data=[...] (use "-full mdat" to show all)
~~~
