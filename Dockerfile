FROM alpine:3.4

RUN apk add --update bash ca-certificates \
  && rm -rf /var/cache/apk/*

COPY cmd/linkchecker /usr/local/bin

ENV TIMEOUT=20

CMD linkchecker -host=$HOST -timeout=$TIMEOUT
