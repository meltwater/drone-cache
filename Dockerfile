# build stage
FROM golang:1.11-alpine AS builder
RUN apk add --update ca-certificates tzdata && update-ca-certificates

# final stage
FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY drone-cache /bin/

ENTRYPOINT ["/bin/drone-cache"]
