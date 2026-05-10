FROM alpine:latest

RUN apk add --no-cache \
    python3 \
    py3-pip \
    go \
    build-base \
    linux-headers \
    time

# Disable cgo. Go builds are faster and we don't need a C compiler for Go code.
ENV CGO_ENABLED=0

WORKDIR /workspace