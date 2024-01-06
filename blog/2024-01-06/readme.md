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

example 2:

~~~
> ffmpeg -i BCRhBaFqtf0.mp4 -c copy -frag_size 6M -movflags dash 6M.mp4
> ffmpeg -v verbose -i 6M.mp4
[AVIOContext @ 000001b858c74c80] Statistics: 360452 bytes read, 9 seeks
~~~

example 3:

~~~
> ffmpeg -i BCRhBaFqtf0.mp4 -c copy -frag_size 7M -movflags dash 7M.mp4
> ffmpeg -v verbose -i 7M.mp4
[AVIOContext @ 000002001d684c80] Statistics: 360452 bytes read, 8 seeks
~~~
