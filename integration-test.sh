#!/bin/bash

trap cleanup 1 2 3 6

cleanup()
{
  echo "stopping containers:"
  docker kill otel-collector
  docker rm otel-collector
  kill "$(jobs -p)"
  exit
}

echo "start otel collector (backend)"
docker run -d --name otel-collector -v "$(pwd)/otel-config.yaml:/etc/otelcol/config.yaml" otel/opentelemetry-collector:0.101.0
# in new terminal
ip=$(docker inspect otel-collector -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}')
#now pick one of the configurations below
echo "build & start metrics-demo"
CGO_ENABLED=0 go build .
OTEL_EXPORTER_OTLP_METRICS_PROTOCOL="http/protobuf" OTEL_EXPORTER_OTLP_METRICS_ENDPOINT="http://$ip:4318/v1/metrics" ./metrics-demo &
sleep 60
echo "increase counter and check"
start=$(docker logs otel-collector 2>&1 | grep "Value:" | tail -n1| awk '{print $2}')
echo "before: $start"
curl "localhost:8080/otlpmetric"
curl "localhost:8080/otlpmetric"
sleep 120 #metrics are only supported once per minute
end=$(docker logs otel-collector 2>&1 | grep "Value:" | tail -n1| awk '{print $2}')
echo "after: $end"
if [ "$end" -gt "$start" ]; then
    echo "test success"
    rm -f metrics-demo
    cleanup
fi

echo "TEST failed"
rm -f metrics-demo
cleanup




