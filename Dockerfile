# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
COPY . .
RUN go build \
	-tags netgo,osusergo \
	-ldflags '-extldflags "-static"' \
	-o /usr/local/bin/fly-autoscaler-multiapp \
	./cmd/fly-autoscaler-multiapp

FROM alpine
COPY --from=builder /usr/local/bin/fly-autoscaler-multiapp /usr/local/bin/fly-autoscaler-multiapp
CMD fly-autoscaler-multiapp -addr :80 -metrics-addr :9090

