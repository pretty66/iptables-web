FROM golang:1.17.8 AS builder
WORKDIR /
COPY . .
RUN make

FROM harbor.oneitfarm.com/bifrost/nettools:v1.1

RUN apt-get update -y && \
    apt-get install iptables -y

WORKDIR /

COPY --from=builder /iptables-server .

ENTRYPOINT ["/iptables-server"]


