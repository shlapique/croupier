# croupier tool🃏
This tool offers files for downloading from disk

## build

```
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags="-s -w" -o croupier .
```
