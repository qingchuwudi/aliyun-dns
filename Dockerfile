FROM golang:1.17.4-alpine
WORKDIR $GOPATH/src/github.com/aliyun-dns
ADD . $GOPATH/src/github.com/aliyun-dns
COPY config.yaml /etc/config.yaml
RUN go build .

ENTRYPOINT  ["./aliyun-dns"]
CMD ["-c", "/etc/config.yaml"]
