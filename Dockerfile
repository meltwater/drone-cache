FROM alpine:3.4

RUN set -ex \
  && apk add --no-cache \
    ca-certificates \
    tar \
  && rm -rf /var/cache/apk/*

ADD drone-s3-cache /bin/
ENTRYPOINT ["/bin/drone-s3-cache"]
