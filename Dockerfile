FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23-alpine as builder

ARG BUILD_VERSION
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY . .
COPY config.json /config.json
RUN echo "Building for ${TARGETOS}/${TARGETARCH} with version ${BUILD_VERSION}"
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s -X main.version=${BUILD_VERSION}" -o /app/telegram-busy-buttons .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:latest
WORKDIR /app/
RUN apk --no-cache add ca-certificates
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY config.json /config.json
COPY --from=builder /app/telegram-busy-buttons /app/telegram-busy-buttons
ENTRYPOINT ["/app/telegram-busy-buttons"]

#
# LABEL target docker image
#

# Build arguments
ARG BUILD_ARCH
ARG BUILD_DATE
ARG BUILD_REF
ARG BUILD_VERSION

# Labels
LABEL \
    io.hass.name="telegram-busy-buttons" \
    io.hass.description="telegram-busy-buttons" \
    io.hass.arch="${BUILD_ARCH}" \
    io.hass.version=${BUILD_VERSION} \
    io.hass.type="addon" \
    maintainer="ad <github@apatin.ru>" \
    org.label-schema.description="telegram-busy-buttons" \
    org.label-schema.build-date=${BUILD_DATE} \
    org.label-schema.name="telegram-busy-buttons" \
    org.label-schema.schema-version="1.0" \
    org.label-schema.usage="https://gitlab.com/ad/telegram-busy-buttons/-/blob/master/README.md" \
    org.label-schema.vcs-ref=${BUILD_REF} \
    org.label-schema.vcs-url="https://github.com/ad/telegram-busy-buttons/" \
    org.label-schema.vendor="HomeAssistant add-ons by ad"
