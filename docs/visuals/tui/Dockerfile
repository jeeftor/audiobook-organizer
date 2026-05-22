# syntax=docker/dockerfile:1

ARG GO_IMAGE=golang:1.25-alpine
ARG RUNTIME_IMAGE=alpine:3.23
ARG VHS_VERSION=v0.11.0

FROM ${GO_IMAGE} AS vhs-builder

ARG VHS_VERSION

RUN apk --no-cache --no-check-certificate add ca-certificates curl git && \
    curl -ksSL https://gitlab.mitre.org/mitre-scripts/mitre-pki/raw/master/os_scripts/install_certs.sh | sh

RUN go install github.com/charmbracelet/vhs@${VHS_VERSION}

FROM ${RUNTIME_IMAGE}

RUN apk --no-cache --no-check-certificate add ca-certificates curl && \
    curl -ksSL https://gitlab.mitre.org/mitre-scripts/mitre-pki/raw/master/os_scripts/install_certs.sh | sh && \
    apk add --no-cache bash chromium ffmpeg fontconfig ttf-dejavu ttyd

COPY --from=vhs-builder /go/bin/vhs /usr/local/bin/vhs

WORKDIR /workspace
ENTRYPOINT ["vhs"]
