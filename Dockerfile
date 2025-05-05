FROM golang:1.24-alpine as mybuilder
WORKDIR $GOPATH/src/github.com/aliyun-dns
ADD . $GOPATH/src/github.com/aliyun-dns
COPY config.yaml /etc/config.yaml
RUN go build .
RUN cp -f ./aliyun-dns /usr/local/bin/aliyun-dns

ENTRYPOINT  ["./aliyun-dns"]
CMD ["-c", "/etc/config.yaml"]

FROM alpine:3.21
COPY --from=mybuilder /usr/local/bin/aliyun-dns /usr/local/bin/aliyun-dns
COPY --from=mybuilder /etc/config.yaml /etc/config.yaml
ENTRYPOINT  ["/usr/local/bin/aliyun-dns"]
CMD ["-c", "/etc/config.yaml"]

