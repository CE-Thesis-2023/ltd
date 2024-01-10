FROM golang:1.21.4-alpine3.18 AS builder

RUN apk add g++ gcc

WORKDIR /build
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY src src

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build \
    -a -installsuffix cgo \
    -ldflags "-w -s" \
    -o main main.go

FROM alpine:3.18 AS runner

RUN apk add ffmpeg

WORKDIR /
COPY configs.json configs.json
COPY --from=builder /build/main main

ENTRYPOINT [ "./main" ]