package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"

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
	router.Static("/static", "./static") // 静的ファイルの読み込み
	router.Use(TracingMiddleware())
	router.LoadHTMLGlob("templates/*") // テンプレートファイルの読み込み

	// Cookieベースのセッションを設定
	secretKey := os.Getenv("SESSION_SECRET_KEY") // Sessionの暗号化キーは固定の値を使用することで、アプリの再起動時にセッションが維持されるようにする
	if secretKey == "" {
		fmt.Println("SESSION_SECRET_KEY環境変数が設定されていません")
		return
	}

	store := cookie.NewStore([]byte(secretKey))
	router.Use(sessions.Sessions("session", store)) // ブラウザのCookieにセッションIDを保存する

	// CSRFミドルウェアの設定
	// HTML内の_csrfの値を取得して、リクエストトークンと比較を行い、一致しない場合ErrorFuncを実行する（https://github.com/utrack/gin-csrf/blob/master/csrf.go）
	router.Use(csrf.Middleware(csrf.Options{
		Secret: secretKey, // 上のCookieベースのセッションと同じ値を指定
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))

	router.GET("/", func(c *gin.Context) {
		token := csrf.GetToken(c)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"head_title": "Golang-Gin",
			"title":      "Main website",
			"content":    "OpenTelemetry with Golang-Gin",
			"csrfToken":  token,
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
