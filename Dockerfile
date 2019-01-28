FROM golang:1.11-alpine as builder

RUN apk add --no-cache make

ADD . /proxy
RUN cd /proxy && make

FROM alpine:latest

WORKDIR /proxy

COPY --from=builder /proxy /usr/local/bin/proxy

RUN chmod +x /usr/local/bin/proxy

EXPOSE 80

ENTRYPOINT ["/usr/local/bin/proxy"]

CMD ["--help"]
