# metrics-demo

Just a basic HTTP Server on port 8080 to provide some prometheus and otlp metrics (good for testing)

The following path are available:

* / returns 200 and Headers
* /ping returns pong
* /q/health/ready readiness check
* /q/health/live liveness check
* /q/metrics metrics
* /metrics on standard path
* / same as '/'
* /otlpmetrics increase otlp metric counter

## run locally

set otel collector http endpoint (if available) :

```bash
docker run -d --name otel-collector -v $(pwd)/otel-config.yaml:/etc/otelcol/config.yaml otel/opentelemetry-collector:0.73.0
docker logs otel-collector -f
# in new terminal
ip=$(docker inspect otel-collector -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}')
#now pick one of the configurations below
OTEL_EXPORTER_OTLP_METRICS_PROTOCOL="http/protobuf" OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://$(echo $ip):4318/v1/metrics go run main.go
OTEL_EXPORTER_OTLP_METRICS_PROTOCOL="http/protobuf" OTEL_EXPORTER_OTLP_ENDPOINT=http://$(echo $ip):4318 go run main.go
OTEL_EXPORTER_OTLP_METRICS_PROTOCOL="grpc" OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://$(echo $ip):4317 go run main.go

```

stop it:

```bash
docker stop otel-collector
docker rm otel-collector
```

For otlp metrics to work endpoint must be running.
