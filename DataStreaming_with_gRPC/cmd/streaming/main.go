package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	_ "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {
	r := gin.Default()
	streaming := r.Group("log/api/v1")
	streaming.Use(TenantidCheck())

	//// gRPC
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint("10.111.1.179:4317"),
		otlptracegrpc.WithInsecure(), // TLSを無効にする場合に指定
	)

	//// HTTP
	// exporter, err := otlptracehttp.New(context.Background(),
	// 	otlptracehttp.WithEndpoint("10.111.1.179:4318"),
	// 	otlptracehttp.WithInsecure(), // TLSを無効にする場合に指定
	// )
	if err != nil {
		log.Fatalln("Failed to set exporter for otlp/grpc")
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(1*time.Second),
			trace.WithMaxExportBatchSize(128),
		),
		trace.WithSampler(trace.AlwaysSample()), // すべてのトレースをサンプリング
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,                          // SchemaURL is the schema URL used to generate the trace ID. Must be set to an absolute URL.
			semconv.ServiceNameKey.String("streaming"), // ServiceNameKey is the key used to identify the service name in a Resource.
		)),
	)
	// プログラム終了時に適切にリソースを解放
	defer tp.Shutdown(context.Background())

	// SetTracerProvider registers `tp` as the global trace provider.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tr := otel.Tracer("streaming")

	streaming.POST("/push", func(c *gin.Context) {
		// ctx, span := tr.Start(context.Background(), "data streaming started")
		_, span := tr.Start(context.Background(), "data push")
		defer span.End()

		// Add attributes to the span
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.remote_addr", c.Request.RemoteAddr),
		)
		fmt.Println("Test Push!")
	})

	r.Run(":8081")
}
