# global sidx

~~~
youtube -b BCRhBaFqtf0 -i 160
~~~

this works:

~~~
bin/mp4split in.mp4
~~~

this fails:

~~~
ffmpeg -i in.mp4 -c copy -frag_size 99K frag.mp4
bin/mp4split frag.mp4
~~~
