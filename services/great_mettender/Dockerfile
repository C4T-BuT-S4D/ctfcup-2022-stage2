FROM golang:1.18.3-alpine

ENV INTERFUCK_PATH /interfuck

# Disables useless warning, not a part of the service.
ENV QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING = true

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /interfuck ./cmd/interfuck/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /gmt_server ./cmd/server/main.go

CMD ["/gmt_server"]
