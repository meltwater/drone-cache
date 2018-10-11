# build stage
FROM golang:1.11.1-alpine AS builder
RUN apk add --update make git

ENV BUILD_DIR /build

COPY go.* Makefile $BUILD_DIR/
WORKDIR $BUILD_DIR
RUN make fetch-dependecies

RUN pwd
RUN ls .
RUN echo "HEDE"

COPY . $BUILD_DIR

RUN pwd
RUN ls .
RUN echo "HEDE"

RUN make drone-s3-cache
RUN cp drone-s3-cache /bin

# final stage
FROM alpine as runner
COPY --from=builder /bin/drone-s3-cache /bin

RUN set -ex \
  && apk add --no-cache \
    ca-certificates \
    tar \
  && rm -rf /var/cache/apk/*

ENTRYPOINT ["/bin/drone-s3-cache"]
