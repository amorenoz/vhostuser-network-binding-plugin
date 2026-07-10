ARG GOLANG_VERSION=1.26
FROM golang:${GOLANG_VERSION} AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM quay.io/centos/centos:stream10-minimal AS runtime

COPY --from=builder /build/build/vhostuser-network-binding-plugin \
    /usr/bin/vhostuser-network-binding-plugin

ENTRYPOINT [ "/usr/bin/vhostuser-network-binding-plugin" ]
