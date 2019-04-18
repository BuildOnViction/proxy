FROM golang:1.11-alpine as builder

RUN apk update \
        && apk upgrade \
        && apk add --no-cache make

RUN mkdir -p $GOPATH/src/github.com/tomochain/proxy
COPY . $GOPATH/src/github.com/tomochain/proxy
RUN cd $GOPATH/src/github.com/tomochain/proxy && make
RUN cp $GOPATH/src/github.com/tomochain/proxy/proxy /usr/local/bin/proxy

FROM alpine:latest

RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    ca-certificates \
    && update-ca-certificates 2>/dev/null || true

WORKDIR /proxy

COPY --from=builder /usr/local/bin/proxy /usr/local/bin/proxy

RUN chmod +x /usr/local/bin/proxy

EXPOSE 3000

CMD ["/usr/local/bin/proxy", "--help"]
