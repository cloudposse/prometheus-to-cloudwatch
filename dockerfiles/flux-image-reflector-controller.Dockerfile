FROM ghcr.io/fluxcd/image-reflector-controller:v0.24.0 as build

FROM artifact.onwalk.net/k8s/alpine-ca:3.13 as prod

RUN apk add --no-cache ca-certificates tini
COPY --from=builder /usr/local/bin/image-reflector-controller /usr/local/bin/
USER 65534:65534

ENTRYPOINT [ "/sbin/tini", "--", "image-reflector-controller" ]
