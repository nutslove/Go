package routers

import (
	"context"
	"ham3/services"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	// OTLPエクスポーターの設定
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(), // TLSを無効にする場合に指定
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	// Tracerの設定
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL, // SchemaURL is the schema URL used to generate the trace ID. Must be set to an absolute URL.
			semconv.ServiceNameKey.String("HAM3"), // ServiceNameKey is the key used to identify the service name in a Resource.
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	v1 := r.Group("/api/v1")

	{
		// LOGaaS関連ルート
		logaas := v1.Group("/logaas")
		{
			logaas.POST("/:logaas_id", services.CreateLogaas)
			logaas.GET("/:logaas_id", services.GetLogaas)
			logaas.DELETE("/:logaas_id", services.DeleteLogaas)
			logaas.GET("/", services.GetLogaases)
		}

		// CaaS関連ルート
		caas := v1.Group("/caas")
		{
			caas.POST("/:caas_id", services.CreateCaas)
			caas.GET("/:caas_id", services.GetCaas)
			caas.DELETE("/:caas_id", services.DeleteCaas)
			caas.GET("/", services.GetCaases)
		}
	}

	// indexページ
	r.GET("/", services.Index)
}
