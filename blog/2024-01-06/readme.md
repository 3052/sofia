# jan 3 2024

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

example 1:

~~~
> ffmpeg -i BCRhBaFqtf0.mp4 -c copy out.mp4
> ffmpeg -v verbose -i out.mp4
[AVIOContext @ 000001e2f20e4c80] Statistics: 226878 bytes read, 2 seeks
~~~

remove `trun`s fail:

~~~
> ffmpeg -i BCRhBaFqtf0.mp4 -c copy -frag_size 6M -movflags dash fail.mp4
> ffmpeg -v verbose -i fail.mp4
[AVIOContext @ 00000244644b4c80] Statistics: 360452 bytes read, 9 seeks
~~~

remove `trun`s pass:

~~~
> ffmpeg -i BCRhBaFqtf0.mp4 -c copy -frag_size 6M pass.mp4
> ffmpeg -v verbose -i pass.mp4
[AVIOContext @ 000002027f2c4c80] Statistics: 262306 bytes read, 8 seeks
~~~
