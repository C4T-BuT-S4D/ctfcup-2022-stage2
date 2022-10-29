.PHONY: all
all:
	go run ./main.go
	zstd -9 large.gz
	base64 <large.gz.zst >large.gz.zst.b64
