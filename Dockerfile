FROM ghcr.io/umputun/baseimage/buildgo:latest as build

ARG GIT_BRANCH
ARG GITHUB_SHA
ARG CI

ADD . /build
WORKDIR /build

RUN go version

RUN \
    if [ -z "$CI" ] ; then \
    echo "runs outside of CI" && version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format=%h)-$(date +%Y%m%dT%H:%M:%S); \
    else version=${GIT_BRANCH}-${GITHUB_SHA:0:7}-$(date +%Y%m%dT%H:%M:%S); fi && \
    echo "version=$version" && \
    cd cmd/bot && go build -o /build/tg-reminder -ldflags "-X main.revision=${version} -s -w"



FROM alpine:3.19

RUN apk add --no-cache tzdata
COPY --from=build /build/tg-reminder /srv/tg-reminder
COPY migrations /srv/db/migrations
RUN \
    adduser -s /bin/sh -D -u 1000 app && chown -R app:app /home/app && \
    chown -R app:app /srv/db && \
    chmod -R 775 /srv/db && \
    ls -la /srv/db && \
    ls -la /srv/db/migrations

USER app
WORKDIR /srv
ENTRYPOINT ["/srv/tg-reminder"]
