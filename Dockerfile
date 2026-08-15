FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}

WORKDIR /src

COPY go.mod go.sum ./
COPY . .

RUN go build \
    -mod=vendor \
    -trimpath \
    -ldflags "-s -w -X 'github.com/yukiteruamano/koma/constant.BuiltAt=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'github.com/yukiteruamano/koma/constant.BuiltBy=docker' -X 'github.com/yukiteruamano/koma/constant.Revision=$(git rev-parse --short HEAD 2>/dev/null || echo docker)'" \
    -o /out/koma .

FROM --platform=$BUILDPLATFORM alpine:3.20 AS runtime-prep

ENV KOMA_DOWNLOADER_PATH=/downloads
ENV KOMA_USER=abc
ENV KOMA_UID=1000
ENV KOMA_GID=1000

WORKDIR /config
RUN mkdir -p "${KOMA_DOWNLOADER_PATH}" && addgroup -g "${KOMA_GID}" "${KOMA_USER}" && adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --ingroup "${KOMA_USER}" \
    --no-create-home \
    --uid "${KOMA_UID}" \
    "${KOMA_USER}" && \
    chown abc:abc /config "${KOMA_DOWNLOADER_PATH}"

FROM alpine:3.20

ENV KOMA_DOWNLOADER_PATH=/downloads
ENV KOMA_USER=abc
ENV KOMA_UID=1000
ENV KOMA_GID=1000

COPY --from=runtime-prep /etc/passwd /etc/passwd
COPY --from=runtime-prep /etc/group /etc/group
COPY --from=runtime-prep /etc/shadow /etc/shadow
COPY --chown=1000:1000 --from=runtime-prep /downloads /downloads
COPY --chown=1000:1000 --from=runtime-prep /config /config
COPY --from=builder /out/koma /usr/local/bin/koma

WORKDIR /config
USER "${KOMA_USER}"
ENTRYPOINT ["/usr/local/bin/koma"]
