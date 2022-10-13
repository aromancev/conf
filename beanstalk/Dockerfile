FROM alpine:3.13

RUN apk add --update --no-cache tini beanstalkd \
    && rm -rf /var/cache/apk/*

ENTRYPOINT ["/sbin/tini","--"]

CMD ["beanstalkd", "-p", "11300", "-b", "/var/lib/beanstalkd"]
