package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	// "go.opentelemetry.io/otel/trace"
)

func main() {
	// Jaeger Exporterの設定
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger:14268/api/traces")))
	if err != nil {
		log.Fatal(err)
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

	router := gin.Default()
	// 静的ファイルのルーティング
	router.Static("/static", "./static")
	router.Use(TracingMiddleware())
	router.LoadHTMLGlob("templates/*")
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"head_title": "Golang-Gin",
			"title":      "Main website",
			"content":    "OpenTelemetry with Golang-Gin",
		})
	})

	router.POST("/login", func(c *gin.Context) {
		// フォームの値を取得
		UserName := c.PostForm("username") // htmlのinputタグのname属性を指定
		Password := c.PostForm("password") // htmlのinputタグのname属性を指定

		// ログイン処理
		if UserName == "admin" && Password == "admin" {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"head_title": "Golang-Gin",
				"title":      "Welcome",
				"content":    "Password Accepted",
			})
		} else {
			c.HTML(403, "index.tmpl", gin.H{
				"head_title": "Golang-Gin",
				"title":      "Login Failed",
				"content":    "Password Rejected",
			})
		}
	})

	router.Run(":8080")
}

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tr := otel.Tracer("gin-server")
		ctx := c.Request.Context()                     // トレースのルートとなるコンテキストを生成
		ctx, span := tr.Start(ctx, c.Request.URL.Path) // トレースの開始 (スパンの開始)
		defer span.End()                               // トレースの終了

		// Add attributes to the span
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.remote_addr", c.Request.RemoteAddr),
		)

		c.Request = c.Request.WithContext(ctx) // コンテキストを更新
		c.Next()
	}
}
