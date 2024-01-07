# jan 3 2024

start with this:

~~~
youtube -b BCRhBaFqtf0 -vc avc1
~~~

run through `mp4ff-info`, below is the smallest result for a file with length in
Windows. remove `trun`s pass:

~~~
> ffmpeg -i in.mp4 -c copy -frag_size 6M -movflags omit_tfhd_offset `
>> omit_tfhd_offset.mp4
> ffmpeg -v verbose -i omit_tfhd_offset.mp4
[AVIOContext @ 000002eadf0b4c80] Statistics: 262306 bytes read, 8 seeks
~~~

below is the smallest result for a file with no length in Windows. remove `trun`s
fail:

~~~
> ffmpeg -i in.mp4 -c copy -frag_size 6M -movflags empty_moov empty_moov.mp4
> ffmpeg -v verbose -i empty_moov.mp4
[AVIOContext @ 0000021ebd954c80] Statistics: 327861 bytes read, 8 seeks
~~~
