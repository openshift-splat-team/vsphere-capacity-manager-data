FROM golang:1.22 AS builder
WORKDIR /go/src/github.com/openshift-splat-team/vsphere-capacity-manager-data
COPY . .
ENV GO_PACKAGE github.com/openshift-splat-team/vsphere-capacity-manager-data
RUN NO_DOCKER=1 make build

FROM registry.access.redhat.com/ubi9-minimal:9.4-949.1716471857
COPY --from=builder /go/src/github.com/openshift-splat-team/vsphere-capacity-manager-data/bin/vcmd /usr/bin/vcmd
ENTRYPOINT ["/usr/bin/vcmd"]
