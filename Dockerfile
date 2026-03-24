FROM golang:1.26-alpine AS builder

RUN	apk add --no-cache \
	bash \
	ca-certificates

COPY . /go/src/avrocat
WORKDIR /go/src/avrocat

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
	git \
	gcc \
	libc-dev \
	libgcc \
	make \
	&& cd /go/src/avrocat \
	&& go mod download \
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
