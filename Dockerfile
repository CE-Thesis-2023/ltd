FROM golang:1.21.4-alpine3.18 AS builder

RUN apk add --no-cache --update gcc g++

WORKDIR /build
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY api api
COPY biz biz
COPY internal internal
COPY helper helper
COPY models models

COPY main.go main.go

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -o main main.go

FROM alpine:3.18 AS runner

WORKDIR /usr/local/bin
COPY /bin/ffmpeg ffmpeg
COPY /bin/ffplay ffplay
COPY /bin/ffprobe ffprobe

WORKDIR /
COPY configs.json configs.json
COPY --from=builder /build/main main

CMD [ "./main" ]