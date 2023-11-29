FROM golang:1.21 as builder

ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV CGO_ENABLED=0

WORKDIR /work
ADD . .
RUN make build-inkd

FROM alpine:3.18 as alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add -U --no-cache ca-certificates tzdata

FROM alpine:3.18
ENV TZ=Asia/Shanghai
ENV INKD_CONFIG=/work/config/inkd.yaml

COPY --from=alpine /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /work/config/inkd.yaml /work/config/inkd.yaml
COPY --from=builder /work/_output/inkd /bin/inkd

WORKDIR /work
CMD ["inkd"]
