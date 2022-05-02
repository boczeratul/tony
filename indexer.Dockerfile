# syntax = docker/dockerfile:experimental
FROM golang:1.17-alpine as compiler
ENV GO111MODULE=on
# cache module
WORKDIR /gopath/src/
COPY go.mod go.sum ./
RUN --mount=type=cache,id=go-mod-cache,target=/go/pkg go mod download

# build release
RUN --mount=target=. \
    --mount=type=cache,id=go-build-cache,target=/root/.cache/go-build \
    --mount=type=cache,id=go-mod-cache,target=/go/pkg \
    CGO_ENABLED=0 go build -v -o /out/main ./indexer/main/main.go

FROM alpine:3.8
RUN addgroup --gid 1001 -S portto && \
    adduser -G portto --shell /bin/false --disabled-password -H --uid 1001 portto
RUN apk add --update --no-cache tzdata
ENV TZ=Asia/Taipei
COPY --from=compiler --chown=portto:portto /out/main /go/bin/main
USER portto
ENTRYPOINT /go/bin/main