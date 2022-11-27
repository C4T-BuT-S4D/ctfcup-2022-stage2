.PHONY: build-local
build-local:
	CGO_ENABLED=0 \
		go build \
			-ldflags="-w -s" \
			-o gmt_checker \
			main.go

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build \
			-ldflags="-w -s" \
			-o gmt_checker \
			main.go

.PHONY: all
all: build
