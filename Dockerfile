# build stage
FROM golang:1.11.5-alpine AS builder
RUN apk add --update make git upx

ENV BUILD_DIR /build

COPY go.* Makefile $BUILD_DIR/
WORKDIR $BUILD_DIR
RUN make fetch-dependecies

COPY . $BUILD_DIR

RUN make drone-s3-cache
RUN make compress
RUN cp drone-s3-cache /bin

# final stage
FROM alpine:3.9 as runner
COPY --from=builder /bin/drone-s3-cache /bin

RUN set -ex \
  && apk add --no-cache \
    ca-certificates \
    tar \
  && rm -rf /var/cache/apk/*

ENTRYPOINT ["/bin/drone-s3-cache"]
