# croupier tool🃏
This tool offers files for downloading from disk

It has prealoding feature with Sliding Window (RingBuffer cache).

![](pics/pic.svg)

## build

```
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags="-s -w" -o croupier .
```
