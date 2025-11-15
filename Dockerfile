FROM golang:1.17.8 AS builder
WORKDIR /app
COPY . .
RUN make

FROM ubuntu:22.04

RUN sed -i s@/archive.ubuntu.com/@/mirrors.cloud.tencent.com/@g /etc/apt/sources.list &&\
    sed -i s@/security.ubuntu.com/@/mirrors.cloud.tencent.com/@g /etc/apt/sources.list &&\
    apt-get update -y &&\
    apt-get install net-tools -y &&\
    apt-get install ca-certificates -y &&\
    apt-get install iptables -y


WORKDIR /app

COPY --from=builder /app/iptables-server .

ENTRYPOINT ["/app/iptables-server"]
