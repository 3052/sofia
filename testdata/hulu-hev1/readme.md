# hulu

~~~
hulu -a hulu.com/watch/0ec7a9e6-d59c-4a73-b9d4-0bb336af58f6
-i 0182f416-7a10-1727-3da6-01000b473bf7

curl -o video.mp4 `
'https://http-fa-darwin.hulustream.com/960/61711960/agave52117579_1000069478253_HEVC10_280_1000069482134_video.mp4?expires=1722137411&keyid=NzkwYmRlZWMzM2VlZDQK&signature=KpOGEHeJ0GsV2Y18j_MsRw1yCNvwmlbR5GAcjk0Cs3c%3D'

bin/mp4split video.mp4
~~~
