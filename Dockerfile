# build stage
FROM golang:1.11.5-alpine AS builder
RUN apk add --update make git upx

ENV BUILD_DIR /build

COPY go.* Makefile $BUILD_DIR/
WORKDIR $BUILD_DIR
RUN make fetch-dependencies

COPY . $BUILD_DIR

RUN make drone-cache
# RUN make compress # TODO: Enable before merge
RUN cp drone-cache /bin

# final stage
FROM alpine:3.9 as runner
COPY --from=builder /bin/drone-cache /bin

# TODO: Remove tar!
RUN set -ex \
  && apk add --no-cache \
    ca-certificates \
    tar \
  && rm -rf /var/cache/apk/*

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/bin/drone-cache"]
