FROM ghcr.io/fluxcd/notification-controller:v0.24.0 as build

FROM artifact.onwalk.net/k8s/alpine-ca:3.13 as prod

LABEL org.opencontainers.image.source="https://github.com/fluxcd/notification-controller"
RUN apk add --no-cache ca-certificates tini
COPY --from=build /usr/local/bin/notification-controller /usr/local/bin/
USER 65534:65534

ENTRYPOINT [ "/sbin/tini", "--", "notification-controller" ]
