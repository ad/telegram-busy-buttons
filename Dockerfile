FROM golang:alpine as builder

ARG BUILD_ARCH

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR $GOPATH/src/app/
COPY . .
COPY config.json /config.json

RUN CGO_ENABLED=0 go build -mod=vendor -ldflags='-w -s -extldflags "-static"' -a -o /go/bin/telegram-busy-buttons .

FROM scratch

ARG BUILD_DATE
ARG BUILD_REF

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/telegram-busy-buttons /go/bin/telegram-busy-buttons
COPY --from=builder /config.json /config.json

EXPOSE 18000

ENTRYPOINT ["/go/bin/telegram-busy-buttons"]

# Labels
LABEL \
    io.hass.name="telegram-busy-buttons" \
    io.hass.description="telegram-busy-buttons" \
    io.hass.arch="${BUILD_ARCH}" \
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