FROM alpine:3.4

RUN apk add --update bash ca-certificates curl \
  && rm -rf /var/cache/apk/*

COPY cmd/linkchecker /usr/local/bin
COPY wait.sh .
COPY .linkignore .

ENV TIMEOUT=20

CMD linkchecker -host=$HOST -timeout=$TIMEOUT
