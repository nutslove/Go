package middleware

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	_ "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func TracingMiddleware() gin.HandlerFunc {
	// // Jaeger Exporterの設定
	// exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger:14268/api/traces")))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// OTLPエクスポーターの設定
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("jaeger:4318"),
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
			semconv.ServiceNameKey.String("Golang-Gin"), // ServiceNameKey is the key used to identify the service name in a Resource.
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func(c *gin.Context) {
		tr := otel.Tracer("gin-server")
		ctx := c.Request.Context()                     // トレースのルートとなるコンテキストを生成
		ctx, span := tr.Start(ctx, c.Request.URL.Path) // トレースの開始 (スパンの開始)
		defer span.End()                               // トレースの終了

		c.Request = c.Request.WithContext(ctx) // コンテキストを更新
		c.Next()                               // 次のミドルウェアを呼び出し // ここでgin.Contextが更新される // この後の処理はgin.Contextの値を参照することができる

		// HTTPステータスコードが400以上の場合、エラーとしてマーク
		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		// Add attributes to the span
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.Int("http.status_code", statusCode),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.remote_addr", c.Request.RemoteAddr),
		)
	}
}
