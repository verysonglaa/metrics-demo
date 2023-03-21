package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	flag "github.com/spf13/pflag"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func promMetric() {

	var (
		opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
			Name: "example_counter",
			Help: "just an example",
		})
	)
	opsProcessed.Inc()

}

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	otlpHandler, shutdownFunc, err := otlpmetric(ctx)
	if err != nil {
		log.Printf("Error creating metric,%s\n", err)
	}
	defer shutdownFunc()

	promMetric()

	http.Handle("/", echo())
	http.Handle("/ping", ping())
	http.Handle("/q/health/ready", ready())
	http.Handle("/q/health/live", live())
	http.Handle("/q/metrics", promhttp.Handler())
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/echo", echo())
	http.Handle("/otlpmetric", otlpHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: http.StripPrefix(os.Getenv("PATH_PREFIX"), http.DefaultServeMux),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("unable to create server: %v\n", err)
			os.Exit(1)
		}
	}()
	log.Println("listen on :8080")

	<-done
	log.Print("server shutting down")

	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed:%+v", err)
	}
	log.Print("server exited properly")
}

func notFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		w.Header().Set("app", "crt-svc-example-service")
		http.Error(w, "Not found", http.StatusNotFound)
	})
}

func ping() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		fmt.Fprintln(w, "pong")
	})
}

func ready() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ready")
	})
}

func live() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "live")
	})
}

func echo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(200)

		host, err := os.Hostname()
		if err == nil {
			fmt.Fprintf(w, "Request served by %s\n\n", host)
		} else {
			fmt.Fprintf(w, "Server hostname unknown: %s\n\n", err.Error())
		}

		fmt.Fprintf(w, "%s %s %s\n", req.Proto, req.Method, req.URL)
		fmt.Fprintln(w, "")

		fmt.Fprintf(w, "Host: %s\n", req.Host)
		for key, values := range req.Header {
			for _, value := range values {
				fmt.Fprintf(w, "%s: %s\n", key, value)
			}
		}

		var body bytes.Buffer
		io.Copy(&body, req.Body) // nolint:errcheck

		if body.Len() > 0 {
			fmt.Fprintln(w, "")
			body.WriteTo(w) // nolint:errcheck
		}
	})
}

func otlpmetric(ctx context.Context) (http.Handler, func(), error) {

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("metrics-demo"),
		semconv.ServiceVersionKey.String("v0.0.0"),
	)
	// Instantiate the OTLP HTTP exporter
	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, nil, err
	}
	// Instantiate the OTLP HTTP exporter, send metrics every minute to OTEL_EXPORTER_OTLP_METRICS_ENDPOINT

	meterProvider := sdk.NewMeterProvider(
		sdk.WithResource(resources),
		sdk.WithReader(sdk.NewPeriodicReader(exporter, sdk.WithInterval(10*time.Second))),
	)

	shutdownFunc := func() {
		err := meterProvider.Shutdown(ctx)
		if err != nil {
			log.Printf("Error on shutdown of meter %s", err)
		}
		log.Println("shutdown of meter")
	}

	// Create an instance on a meter for the given instrumentation scope
	meter := meterProvider.Meter(
		"github.com/verysonglaa/metrics-demo",
		metric.WithInstrumentationVersion("v0.0.0"),
	)

	// Create counter
	requestCount, err := meter.Int64Counter(
		"request_count",
		instrument.WithDescription("Incoming request count"),
		instrument.WithUnit("request"),
	)
	if err != nil {
		log.Printf("Could not send metric to exporter (%s)", err)
	}
	//set to 1 on initialising
	requestCount.Add(ctx, 1)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Record measurements
		attrs := semconv.HTTPServerMetricAttributesFromHTTPRequest("", req)
		requestCount.Add(ctx, 1, attrs...)
		log.Print("added 1 to otlp counter")
		echo()
	}), shutdownFunc, nil
}
