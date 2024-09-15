# global sidx

https://github.com/Eyevinn/mp4ff/issues/311

start with this:

~~~
youtube -b BCRhBaFqtf0 -i 137
~~~

sanity check:

~~~
> ffmpeg -v verbose -i in.mp4
[AVIOContext @ 0000016462328bc0] Statistics: 98304 bytes read, 0 seeks
~~~

then:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 99K frag.mp4
~~~

then:

~~~
> ffmpeg -v verbose -i frag.mp4
[AVIOContext @ 000002507f688bc0] Statistics: 14065552 bytes read, 411 seeks
~~~

then:

~~~
bin/mp4split frag.mp4
~~~
