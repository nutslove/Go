package routers

import (
	"context"
	"fmt"
	"log"

	"ham3/middlewares"
	"ham3/services"
	"ham3/utilities"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Namespaceを作成するマニフェストの定義
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-namespace",
		},
	}

	// NamespaceをKubernetesクラスターに適用
	_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating namespace: %v\n", err)
		// return
	}

	fmt.Println("Namespace created successfully")

	v1 := r.Group("/api/v1")

	{
		// CaaS関連ルート
		caas := v1.Group("/caas")
		{
			caas.Use(middlewares.TracerSetting("CaaS"))
			caas.POST("/:caas_id", services.CreateCaas)
			caas.GET("/:caas_id", services.GetCaas)
			caas.DELETE("/:caas_id", services.DeleteCaas)
			caas.GET("/", services.GetCaases)
		}

		// LOGaaS関連ルート
		logaas := v1.Group("/logaas")
		{
			logaas.Use(middlewares.TracerSetting("LOGaaS"))
			logaas.POST("/:logaas_id", services.CreateLogaas)
			logaas.GET("/:logaas_id", services.GetLogaas)
			logaas.DELETE("/:logaas_id", services.DeleteLogaas)
			logaas.GET("/", services.GetLogaases)
		}
	}

	// indexページ
	r.GET("/", services.Index)
}
