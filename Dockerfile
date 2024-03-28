FROM alpine:3.19
COPY metrics-demo /usr/local/bin/
RUN chown 65534 /usr/local/bin/metrics-demo 
RUN apk add --update \
    curl \
    && rm -rf /var/cache/apk/*
ENTRYPOINT ["/usr/local/bin/metrics-demo"]
USER 65534