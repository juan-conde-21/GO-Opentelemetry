package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
	meter  metric.Meter
)

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("http-server"),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

func initMeter(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second))

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("http-server"),
		)),
	)

	return mp, nil
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "helloHandler")
	defer span.End()

	fmt.Fprintf(w, "Hello, World!")
	log.Println("Handled /hello request")
	span.AddEvent("Handled /hello request")
}

func sumHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "sumHandler")
	defer span.End()

	query := r.URL.Query()
	num1Str := query.Get("num1")
	num2Str := query.Get("num2")

	num1, err1 := strconv.Atoi(num1Str)
	num2, err2 := strconv.Atoi(num2Str)

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	result := num1 + num2
	fmt.Fprintf(w, "Sum: %d", result)
	log.Printf("Handled /sum request: %d + %d = %d", num1, num2, result)
	span.AddEvent(fmt.Sprintf("Handled /sum request: %d + %d = %d", num1, num2, result))
}

func subtractHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "subtractHandler")
	defer span.End()

	query := r.URL.Query()
	num1Str := query.Get("num1")
	num2Str := query.Get("num2")

	num1, err1 := strconv.Atoi(num1Str)
	num2, err2 := strconv.Atoi(num2Str)

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	result := num1 - num2
	fmt.Fprintf(w, "Subtraction: %d", result)
	log.Printf("Handled /subtract request: %d - %d = %d", num1, num2, result)
	span.AddEvent(fmt.Sprintf("Handled /subtract request: %d - %d = %d", num1, num2, result))
}

func main() {
	ctx := context.Background()

	tp, err := initTracer(ctx)
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() { _ = tp.Shutdown(ctx) }()

	mp, err := initMeter(ctx)
	if err != nil {
		log.Fatalf("failed to initialize meter: %v", err)
	}
	defer func() { _ = mp.Shutdown(ctx) }()

	tracer = tp.Tracer("http-server")
	meter = mp.Meter("http-server")

	requestCounter, err := meter.Int64Counter(
		"http_server_requests_total",
		metric.WithDescription("Total number of HTTP requests received"),
	)
	if err != nil {
		log.Fatalf("failed to create request counter: %v", err)
	}

	helloHandlerWithTelemetry := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestCounter.Add(ctx, 1)
		helloHandler(w, r)
	}), "helloHandler")

	sumHandlerWithTelemetry := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestCounter.Add(ctx, 1)
		sumHandler(w, r)
	}), "sumHandler")

	subtractHandlerWithTelemetry := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestCounter.Add(ctx, 1)
		subtractHandler(w, r)
	}), "subtractHandler")

	http.Handle("/hello", helloHandlerWithTelemetry)
	http.Handle("/sum", sumHandlerWithTelemetry)
	http.Handle("/subtract", subtractHandlerWithTelemetry)

	fmt.Println("Server is listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
