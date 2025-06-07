FROM golang:alpine as builder
MAINTAINER Helton Marques <hmarques@themeetgroup.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	bash \
	ca-certificates

COPY . /go/src/github.meetmecorp.com/hmarques/avrocat

RUN set -x \
	&& apk add --no-cache --virtual .build-deps make \
	&& cd /go/src/github.meetmecorp.com/hmarques/avrocat \
	&& make \
	&& mv avrocat /usr/bin/avrocat \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM alpine:latest

COPY --from=builder /usr/bin/avrocat /usr/bin/avrocat
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "avrocat" ]
CMD [ "--help" ]
