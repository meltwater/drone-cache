# build stage
FROM golang:1.11-alpine AS builder
RUN apk add --update make git upx ca-certificates \
  && update-ca-certificates

RUN adduser -D -g '' appuser

ENV BUILD_DIR /build

COPY go.* Makefile $BUILD_DIR/
WORKDIR $BUILD_DIR
RUN make fetch-dependencies

COPY . $BUILD_DIR

# RUN make drone-cache
RUN make build-compressed
RUN cp drone-cache /bin

# final stage
FROM alpine:3.9 as runner
RUN apk add --no-cache su-exec

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /bin/drone-cache /bin

COPY scripts/entrypoint.sh /bin/entrypoint.sh
RUN chmod +x /bin/entrypoint.sh

# RUN mkdir -m 755 /drone

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD ["/bin/drone-cache"]
