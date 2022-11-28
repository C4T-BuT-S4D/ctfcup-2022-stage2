FROM golang:latest AS build

WORKDIR /build
ADD . .
RUN CGO_ENABLED=0 go build -o proxy .

FROM scratch
COPY --from=build /build/proxy /usr/local/proxyz/app

WORKDIR /usr/local/proxyz
ENTRYPOINT ["/usr/local/proxyz/app"]