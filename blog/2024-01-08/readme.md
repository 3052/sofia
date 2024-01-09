# Jan 8 2024

two:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 39M -movflags empty_moov two.mp4
~~~

one:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 49M -movflags empty_moov one.mp4
~~~
