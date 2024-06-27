package routers

import (
	"context"
	"log"
	"net/http"

	"ham3/middlewares"
	"ham3/models"
	"ham3/services"
	"ham3/utilities"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"k8s.io/client-go/kubernetes"
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

	// NewTracerProviderを複数設定しても、最後の設定が有効になる
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,                     // SchemaURL is the schema URL used to generate the trace ID. Must be set to an absolute URL.
			semconv.ServiceNameKey.String("HAM3"), // ServiceNameKey is the key used to identify the service name in a Resource.
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	config, err := utilities.GetKubeconfig()
	if err != nil {
		log.Fatalf("Failed to get kubeconfig: %v", err)
	}

	// Kubernetesクライアントの作成
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// DB接続
	db := models.ConnectDb()

	v1 := r.Group("/api/v1")

	// HeaderにTokenが存在するかチェック
	v1.Use(middlewares.CheckTokenExists())

	{
		// CaaS関連ルート
		caas := v1.Group("/caas")
		{
			caas.Use(middlewares.TracerSetting("CaaS"))
			caas.POST("/:caas_id", func(c *gin.Context) { services.CreateCaas(c.Request.Context(), c, clientset, db) })
			caas.GET("/:caas_id", func(c *gin.Context) { services.GetCaas(c.Request.Context(), c, clientset, db) })
			caas.DELETE("/:caas_id", func(c *gin.Context) { services.DeleteCaas(c.Request.Context(), c, clientset, db) })
			caas.GET("/", func(c *gin.Context) { services.GetCaases(c.Request.Context(), c, clientset, db) })
		}

		// LOGaaS関連ルート
		logaas := v1.Group("/logaas")
		{
			logaas.Use(middlewares.TracerSetting("LOGaaS"))
			logaas.POST("/:logaas_id", func(c *gin.Context) { services.CreateLogaas(c.Request.Context(), c, clientset, db) })
			logaas.GET("/:logaas_id", func(c *gin.Context) { services.GetLogaas(c.Request.Context(), c, clientset, db) })
			logaas.PUT("/:logaas_id", func(c *gin.Context) { services.UpdateLogaas(c.Request.Context(), c, clientset, db) })
			logaas.DELETE("/:logaas_id", func(c *gin.Context) { services.DeleteLogaas(c.Request.Context(), c, clientset, db) })
			logaas.GET("/", func(c *gin.Context) { services.GetLogaases(c.Request.Context(), c, clientset, db) })
		}
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9090", nil)
	}()

	// indexページ
	r.GET("/", services.Index)
}
