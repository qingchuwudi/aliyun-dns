FROM golang:1.15-alpine
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY="https://goproxy.cn,https://goproxy.io,direct"

WORKDIR $GOPATH/src/github.com/aliyun-dns
ADD . $GOPATH/src/github.com/aliyun-dns
COPY config.yaml /etc/config.yaml
RUN go build .

ENTRYPOINT  ["./aliyun-dns"]
CMD ["-c", "/etc/config.yaml"]
