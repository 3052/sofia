# jan 3 2024

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

if you do this:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 6M frag.mp4
~~~

Windows reports as 1m10s, FFmpeg and MPC-HC reports as 8m1s. If you remove the
`trun` boxes, then everything reports as 1m10s.

if you do this:

~~~
ffmpeg -i frag.mp4 -c copy out.mp4
~~~

everything reports as 8m1s. resulting file has no `trun` boxes. can we kill
these:

~~~
[stss] size=444 version=0 flags=000000
- syncSampleCount: 107

[ctts] size=101808 version=0 flags=000000
- sampleCount: 12724
~~~
