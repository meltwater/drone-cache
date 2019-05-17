# build stage
FROM golang:1.11-alpine AS builder
RUN apk add --update make upx git ca-certificates tzdata && update-ca-certificates

ADD . /opt
WORKDIR /opt
RUN make build-compressed # CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# final stage
FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /opt/drone-cache /bin/drone-cache

ENTRYPOINT ["/bin/drone-cache"]
