FROM golang:1.11-alpine as builder

RUN apk add --no-cache make

RUN mkdir -p ${HOME}/go/src/github.com/tomochain/proxy
ADD . ${HOME}/go/src/github.com/tomochain/proxy
RUN cd ${HOME}/go/src/github.com/tomochain/proxy && make
RUN cp ${HOME}/go/src/github.com/tomochain/proxy/proxy /usr/local/bin/proxy
RUN chmod +x /usr/local/bin/proxy

FROM alpine:latest

WORKDIR /proxy

COPY --from=builder /usr/local/bin/proxy /usr/local/bin/proxy

RUN chmod +x /usr/local/bin/proxy

EXPOSE 80

ENTRYPOINT ["/usr/local/bin/proxy"]

CMD ["--help"]
