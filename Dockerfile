# build stage
FROM golang:alpine AS builder
RUN apk add --update make curl git

ENV SRC ${GOPATH}/src/github.com/kakkoyun/drone-s3-cache

COPY glide.* Makefile $SRC/
WORKDIR $SRC
RUN make fetch-dependecies

ADD . $SRC
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
