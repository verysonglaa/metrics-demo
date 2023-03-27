FROM alpine:3.15.4
COPY metrics-demo /usr/local/bin/
RUN chown 65534 /usr/local/bin/metrics-demo 
ENTRYPOINT ["/usr/local/bin/metrics-demo"]
USER 65534