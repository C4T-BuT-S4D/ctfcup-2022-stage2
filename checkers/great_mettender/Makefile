.PHONY: build-local
build-local:
	CGO_ENABLED=0 \
		go build \
			-ldflags="-w -s" \
			-o checker.py \
			main.go

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build \
			-ldflags="-w -s" \
			-o checker.py \
			main.go

.PHONY: compress
compress:
	upx --lzma -9 checker.py

.PHONY: all
all: build compress
