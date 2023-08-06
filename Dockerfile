ARG TARGET_DIR=/app
ARG GOBIN=/.bin

FROM golang:1.20.0-alpine as builder
ARG TARGET_DIR
ARG GOBIN
ARG SSH_PRIVATE_KEY
RUN apk add --update make git mercurial openssh
RUN mkdir -p ~/.ssh && umask 0077 && echo "${SSH_PRIVATE_KEY}" | base64 -d > ~/.ssh/id_rsa \
  && git config --global url."git@bitbucket.org:".insteadOf https://bitbucket.org/ \
  && git config --global url."git@github.com:".insteadOf https://github.com/ \
  && ssh-keyscan bitbucket.org >> ~/.ssh/known_hosts \
  && ssh-keyscan github.com >> ~/.ssh/known_hosts

ENV GO111MODULE=on
ENV GOPRIVATE $GOPRIVATE
WORKDIR /go/src/github.com/levongh/profile

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN GOBIN=$GOBIN make install-tools
RUN TARGET_DIR=$TARGET_DIR make build

# Dev image
FROM builder AS dev
EXPOSE 8030
ENTRYPOINT make watch

# Run
FROM alpine:3.12
ARG TARGET_DIR
ARG GOBIN
RUN apk add --no-cache ca-certificates make git \
    && rm -rf /var/cache/apk/*

WORKDIR /app/
ENV PORT=8030

COPY --from=builder ${TARGET_DIR}/profile profile
COPY --from=builder ${GOBIN}/migrate migrate
COPY --from=builder /go/src/github.com/levongh/profile/Makefile Makefile
COPY --from=builder /go/src/github.com/levongh/profile/db/migrations db/migrations
EXPOSE ${PORT}
