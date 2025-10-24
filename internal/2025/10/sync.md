# sync

~~~
ffmpeg `
-i 460ddcbc-1fcb-442a-918a-15835c9bb683_2025-10-23_17-16-12.mp4 `
-i 460ddcbc-1fcb-442a-918a-15835c9bb683_2025-10-23_17-16-12.eng.m4a `
-c copy `
-movflags negative_cts_offsets `
-use_editlist 0 `
.mp4
~~~

- https://code.ffmpeg.org/FFmpeg/FFmpeg/issues/20742
- https://github.com/BtbN/FFmpeg-Builds/issues/548
- https://github.com/GyanD/codexffmpeg/issues/203
- https://trac.ffmpeg.org/ticket/11034
