package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	_ "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var router *gin.Engine

func main() {
	PostgresUser := os.Getenv("POSTGRES_USER")
	PostgresPassword := os.Getenv("POSTGRES_PASSWORD")
	PostgresDatabase := os.Getenv("POSTGRES_DB")

	dsn := "port=5432 sslmode=disable TimeZone=Asia/Tokyo host=postgres" + " user=" + PostgresUser + " password=" + PostgresPassword + " dbname=" + PostgresDatabase
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// セッションを利用するために、DBに接続する
	sqldb, err := db.DB()
	// 接続エラー時の処理
	if err != nil {
		panic(err)
	}
	// DB接続解除
	defer sqldb.Close()

	sqldb.SetMaxIdleConns(10)           // SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqldb.SetMaxOpenConns(100)          // SetMaxOpenConns sets the maximum number of open connections to the database.
	sqldb.SetConnMaxLifetime(time.Hour) // SetConnMaxLifetime sets the maximum amount of time a connection may be reused.

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

	router = gin.Default()
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
			AfterLogin(c)
		} else {
			c.HTML(http.StatusUnauthorized, "index.tmpl", gin.H{
				"head_title": "Golang-Gin",
				"title":      "Login Failed",
				"content":    "Password Rejected",
			})
		}
	})

	router.Run(":8080")
}

func AfterLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "main.tmpl", gin.H{
		"head_title": "seeds_candle",
		"title":      "Hope you spend a good time!",
		"content":    "Happiness is a choice.",
	})
}

func TracingMiddleware() gin.HandlerFunc {
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
