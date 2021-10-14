FROM golang:alpine AS builder

WORKDIR /data

ADD . /data

RUN CGO_ENABLED=0 go build -o hb-press-bot .

FROM alpine 

COPY --from=builder /data/hb-press-bot /hb-press-bot

RUN apk --no-cache add ca-certificates

ENTRYPOINT ["/hb-press-bot"]
