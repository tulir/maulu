FROM golang:1-alpine AS builder

RUN apk add --no-cache git
RUN wget -qO /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64
RUN chmod +x /usr/local/bin/dep

COPY Gopkg.lock Gopkg.toml /go/src/maunium.net/go/maulu/
WORKDIR /go/src/maunium.net/go/maulu
RUN dep ensure -vendor-only

COPY . /go/src/maunium.net/go/maulu
RUN CGO_ENABLED=0 go build -o /usr/bin/maulu


FROM scratch

COPY --from=builder /usr/bin/maulu /usr/bin/maulu

CMD ["/usr/bin/maulu"]
