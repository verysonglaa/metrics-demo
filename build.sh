#!/bin/bash
CGO_ENABLED=0 go build .
docker build -t grafgabriel/metrics-demo:0.0.2 --no-cache .
docker push grafgabriel/metrics-demo:0.0.2
