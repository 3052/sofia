# jan 3 2024

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

fragment it:

~~~
ffmpeg -i BCRhBaFqtf0.mp4 -frag_size 9999 -c copy frag.mp4
~~~

split it:

~~~
mp4split frag.mp4
~~~

join it:

~~~
go run join.go
~~~

combine fragments:

~~~
ffmpeg -i frag.mp4 -c copy out.mp4
~~~
