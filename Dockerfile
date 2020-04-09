# This file is designed to only used by goreleaser.
FROM golang:1.14.1-alpine3.11 AS builder
RUN apk add --update --no-cache ca-certificates tzdata && update-ca-certificates

FROM scratch as runner

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY drone-cache /bin/drone-cache

LABEL vendor="meltwater" \
    name="drone-cache" \
    description="A Drone plugin for caching current workspace files between builds to reduce your build times." \
    maintainer="Kemal Akkoyun <kakkoyun@gmail.com>" \
    org.label-schema.description="A Drone plugin for caching current workspace files between builds to reduce your build times." \
    org.label-schema.docker.cmd="docker run --rm -v '$(pwd)':/app -e DRONE_REPO=octocat/hello-world -e DRONE_REPO_BRANCH=master -e DRONE_COMMIT_BRANCH=master -e PLUGIN_MOUNT=/app/node_modules -e PLUGIN_RESTORE=false -e PLUGIN_REBUILD=true -e PLUGIN_BUCKET=<bucket> -e AWS_ACCESS_KEY_ID=<token> -e AWS_SECRET_ACCESS_KEY=<secret> meltwater/drone-cache" \
    org.label-schema.vcs-url="https://github.com/meltwater/drone-cache" \
    org.label-schema.vendor="meltwater" \
    org.label-schema.usage="https://underthehood.meltwater.com/drone-cache" \
    org.opencontainers.image.authors="Kemal Akkoyun <kakkoyun@gmail.com>" \
    org.opencontainers.image.url="https://github.com/meltwater/drone-cache" \
    org.opencontainers.image.documentation="https://underthehood.meltwater.com/drone-cache" \
    org.opencontainers.image.source="https://github.com/meltwater/drone-cache/blob/master/Dockerfile" \
    org.opencontainers.image.vendor="meltwater" \
    org.opencontainers.image.licenses="Apache-2.0" \
    org.opencontainers.image.title="drone-cache" \
    org.opencontainers.image.description="A Drone plugin for caching current workspace files between builds to reduce your build times."

ENTRYPOINT ["/bin/drone-cache"]


