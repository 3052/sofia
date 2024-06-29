# max

~~~
> max -a play.max.com/video/watch/b3b1410a-0c85-457b-bcc7-e13299bea2a8/1623fe4c-ef6e-4dd1-a10c-4a181f5f6579
2024/06/29 17:12:33 INFO GET URL="https://akm.prd.media.h264.io/r/dash.mpd?f.audioTrack=en-US%7Cprogram&f.audioTrack=es-419%7Cprogram&f.videoCodec=avc&f.videoDynamicRange=sdr&r.duration=4.004000&r.duration=5571.816250&r.keymod=2&r.main=1&r.manifest=0cca0350-b0d9-47bb-98d5-d5d81a73fee5%2F1_f40015.mpd&r.manifest=59da086b-1d1e-48fa-b318-782408318b54%2F0_c3ecd7.mpd&r.origin=cfc%7Cprd-wbd-amer-vod&x-wbd-tenant=beam&x-wbd-user-home-market=amer"
bandwidth = 258322
codecs = ec-3
type = audio/mp4
lang = en-US
period = 1 2 3 4 5 6
id = a2
~~~

result:

~~~
curl -O https://akm.prd.media.h264.io/59da086b-1d1e-48fa-b318-782408318b54/a/0_fa3c08/a2.mp4
mp4split a2.mp4
~~~
