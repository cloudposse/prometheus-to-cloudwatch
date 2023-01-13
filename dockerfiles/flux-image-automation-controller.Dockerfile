FROM ghcr.io/fluxcd/image-automation-controller:v0.24.0 as build

FROM artifact.onwalk.net/alpine-ca:3.13

# Copy over binary from build
COPY --from=build /image-automation-controller /usr/local/bin/

USER 65534:65534
ENTRYPOINT [ "image-automation-controller" ]
