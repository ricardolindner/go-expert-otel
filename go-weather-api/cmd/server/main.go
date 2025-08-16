package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ricardolindner/go-expert-otel/goweather-api/internal/handlers"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

func initTracer() *sdktrace.TracerProvider {
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint("otel-collector:4317"))
	if err != nil {
		log.Fatalf("failed to create OTLP exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("go-weather-api"),
		)),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp
}

func main() {
	tp := initTracer()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /weather", handlers.GetWeather)

	handler := otelhttp.NewHandler(mux, "go-weather-api-http-request")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
