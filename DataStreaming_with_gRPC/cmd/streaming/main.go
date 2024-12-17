package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	_ "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	_ "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {
	r := gin.Default()
	streaming := r.Group("log/api/v1")
	streaming.Use(TenantidCheck())

	// #### Trace関連設定
	/// gRPC
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint("10.111.1.179:4317"),
		otlptracegrpc.WithInsecure(), // TLSを無効にする場合に指定
	)
	if err != nil {
		log.Fatalln("Failed to set exporter for otlp/grpc")
	}

	/// HTTP
	// exporter, err := otlptracehttp.New(context.Background(),
	// 	otlptracehttp.WithEndpoint("10.111.1.179:4318"),
	// 	otlptracehttp.WithInsecure(), // TLSを無効にする場合に指定
	// )
	// if err != nil {
	// 	log.Fatalln("Failed to set exporter for otlp/http")
	// }

	// リソース属性の設定
	// resource, err := resource.New(context.Background(),
	// 	resource.WithSchemaURL(semconv.SchemaURL),
	// 	resource.WithAttributes(
	// 		semconv.ServiceNameKey.String("streaming"),
	// 		semconv.ServiceVersionKey.String("1.0.0"),
	// 	),
	// )
	if err != nil {
		fmt.Errorf("failed to create resource: %w", err)
	}
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,                          // SchemaURL is the schema URL used to generate the trace ID. Must be set to an absolute URL.
		semconv.ServiceNameKey.String("streaming"), // ServiceNameKey is the key used to identify the service name in a Resource.
		semconv.ServiceVersionKey.String("1.0.1"),
	)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(1*time.Second),
			trace.WithMaxExportBatchSize(128),
		),
		trace.WithSampler(trace.AlwaysSample()), // すべてのトレースをサンプリング
		trace.WithResource(resource),
	)
	// プログラム終了時に適切にリソースを解放
	defer tp.Shutdown(context.Background())

	// SetTracerProvider registers `tp` as the global trace provider.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tr := otel.Tracer("streaming")

	// #### Metric関連設定
	ctxmetric := context.Background()
	promExporter, err := otlpmetricgrpc.New(ctxmetric,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
	)

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(promExporter,
			// Default is 1m
			sdkmetric.WithInterval(3*time.Second))),
	)
	defer func() {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown meter provider: %v", err)
		}
	}()

	otel.SetMeterProvider(meterProvider)
	meter := otel.Meter("streaming")
	// ヒストグラムの作成
	histogram, err := meter.Float64Histogram(
		"request_duration_seconds",
		metric.WithDescription("Request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Fatal(err)
	}

	streaming.POST("/push", func(c *gin.Context) {
		// 処理時間の計測
		startTime := time.Now()
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
		span.AddEvent("data push executed")

		fmt.Println("Test Push!")
		duration := time.Since(startTime).Seconds()
		// exemplar付きでメトリクスを記録
		// トレースIDとスパンIDを属性として追加
		histogram.Record(ctxmetric, duration,
			metric.WithAttributes(
				attribute.String("trace_id", span.SpanContext().TraceID().String()),
				attribute.String("span_id", span.SpanContext().SpanID().String()),
			),
		)
		fmt.Println("trace-id:", span.SpanContext().TraceID().String())
		fmt.Println("span-id:", span.SpanContext().SpanID().String())
	})

	r.Run(":8081")
}
