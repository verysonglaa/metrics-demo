#!/bin/bash
CGO_ENABLED=0 go build .
docker build -t grafgabriel/metrics-demo:0.0.4 --no-cache .
docker push grafgabriel/metrics-demo:0.0.4
