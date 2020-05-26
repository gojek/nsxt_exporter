FROM quay.io/prometheus/busybox:latest
LABEL maintainer="Giri Kuncoro <girikuncoro@gmail.com>"
LABEL maintainer="William Albertus Dembo <w.albertusd@gmail.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/nsxt_exporter /bin/nsxt_exporter

EXPOSE      9744
USER        nobody
ENTRYPOINT  [ "/bin/nsxt_exporter" ]