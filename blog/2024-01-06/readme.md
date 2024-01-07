# jan 3 2024

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

if you do this:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 6M out.mp4
~~~

Windows reports as 1m10s, FFmpeg and MPC-HC reports as 8m1s. If you remove the
`trun` boxes, then everything reports as 1m10s.

---

run through `mp4ff-info`, below is the smallest result for a file with length in
Windows:

~~~
> ffmpeg -i in.mp4 -c copy -frag_size 6M -movflags omit_tfhd_offset `
>> omit_tfhd_offset.mp4
> ffmpeg -v verbose -i omit_tfhd_offset.mp4
[AVIOContext @ 000002eadf0b4c80] Statistics: 262306 bytes read, 8 seeks
~~~

below is the smallest result for a file with no length in Windows:

~~~
> ffmpeg -i in.mp4 -c copy -frag_size 6M -movflags empty_moov empty_moov.mp4
> ffmpeg -v verbose -i empty_moov.mp4
[AVIOContext @ 0000021ebd954c80] Statistics: 327861 bytes read, 8 seeks
~~~
