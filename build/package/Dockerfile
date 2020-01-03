# Build
ARG GO_VERSION=1.13.4
FROM golang:${GO_VERSION}-alpine AS build
ARG BIN_VERSION
ARG GIT_COMMIT
ARG GIT_SHA
ARG GIT_TAG
ARG GIT_DIRTY
ARG GOPROXY
ENV GOPROXY=$GOPROXY
ENV BIN_OUTDIR=/
ENV BIN_NAME=sslcheck
RUN apk add --update-cache \
		build-base \
		git \
		make \
		upx
COPY ./ /go/sslcheck
WORKDIR /go/sslcheck
RUN make build
RUN upx -9 $BIN_OUTDIR/$BIN_NAME

# Image
FROM alpine:3.11
RUN apk add --update-cache ca-certificates
COPY --from=build /sslcheck /usr/local/bin
ENTRYPOINT ["sslcheck"]
