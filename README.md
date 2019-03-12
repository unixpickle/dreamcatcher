# dreamcatcher

Proxy a remote file that can be accessed via HTTP Byte-Ranges and cache the contents in RAM.

For example, suppose you want to stream a video file at `http://foo.com/movie.mp4`. You can proxy it with dreamcatcher like so:

```
$ dreamcatcher http://foo.com/movie.mp4
2019/03/12 18:08:39 Using filename: movie.mp4
2019/03/12 18:08:39 Listening on address :8080 ...
```

Now you can go to `http://localhost:8080` to acces the file. If you stream this file with VLC, then it will be cached in RAM and access to it should become very fast.
