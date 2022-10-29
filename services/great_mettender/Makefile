.PHONY: lint-proto
lint-proto:
	cd proto && buf lint && buf build

.PHONY: proto
proto: lint-proto
	cd proto && buf generate

.PHONY: tidy
tidy:
	go mod tidy -compat="1.18"

.PHONY: lint
lint: lint-proto

.PHONY: test-go
test-go:
	go test -race ./...

.PHONY: test
test: test-go

.PHONY: bf-hello
bf-hello:
	go run ./cmd/interfuck testdata/hello.bf.gz.zst.b64 testdata/input.gz.zst.b64

.PHONY: bf-cat
bf-cat:
	go run ./cmd/interfuck testdata/cat1.bf.gz.zst.b64 testdata/input.gz.zst.b64

