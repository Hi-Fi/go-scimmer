FROM alpine:latest as certificates
RUN apk --update add ca-certificates

FROM golang:1.15 as builder
WORKDIR /go/src/app
ADD . .

RUN make build-linux-amd64

FROM scratch
WORKDIR /code
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/app/bin/linux/go-scimmer /usr/bin/go-scimmer
WORKDIR /data
VOLUME /data
USER 1001

ENTRYPOINT ["/usr/bin/go-scimmer"]
