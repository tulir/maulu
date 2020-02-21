FROM golang:1-alpine AS builder

RUN apk add --no-cache git
WORKDIR /build/maulu
COPY . /build/maulu
RUN go build -o /usr/bin/maulu

FROM alpine

COPY --from=builder /usr/bin/maulu /usr/bin/maulu

CMD ["/usr/bin/maulu"]
