# metrics-demo

Just a basic HTTP Server on port 8080 to provide some metrics (good for testing)

The following path are available:
* / returns 200 and Headers
* /ping returns pong
* /q/health/ready readiness check
* /q/health/live liveness check
* /q/metrics metrics
* /metrics on standard path
* / same as '/'