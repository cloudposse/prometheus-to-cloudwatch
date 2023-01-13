FROM ghcr.io/fluxcd/image-automation-controller:v0.24.0 as build

FROM artifact.onwalk.net/k8s/alpine-ca:3.13 as prod
ARG TARGETPLATFORM
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf
COPY --from=build /usr/local/bin/image-automation-controller /usr/local/bin/

USER 65534:65534
ENTRYPOINT [ "image-automation-controller" ]
