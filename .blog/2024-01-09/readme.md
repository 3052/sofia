# Jan 9 2024

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

then:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 9K -movflags empty_moov frag.mp4
~~~

check this out. if I take a "normal" file:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 9K frag.mp4
~~~

you get the poor result of tools reading nearly the entire file:

~~~
> ffmpeg -v verbose -i frag.mp4
[AVIOContext @ 000001997f374c80] Statistics: 38141952 bytes read, 48 seeks
~~~

but if you use this undocumented option:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 9K -movflags global_sidx frag.mp4
~~~

then the amount read drops by 99%:

~~~
> ffmpeg -v verbose -i frag.mp4
[AVIOContext @ 000002128d154c80] Statistics: 131076 bytes read, 3 seeks
~~~

https://github.com/FFmpeg/FFmpeg/blob/master/libavformat/movenc.c

does this module have a similar option? this works:

~~~
> ffmpeg -i in.mp4 -c copy -frag_size 99K -movflags dash dash.mp4
> ffmpeg -v verbose -i dash.mp4
[AVIOContext @ 00000221423e4c80] Statistics: 14155780 bytes read, 413 seeks

> ffmpeg -i in.mp4 -c copy -frag_size 99K -movflags dash+global_sidx dash.mp4
> ffmpeg -v verbose -i dash.mp4
[AVIOContext @ 000001dc9c144c80] Statistics: 131076 bytes read, 2 seeks
~~~
