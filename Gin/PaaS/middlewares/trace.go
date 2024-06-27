package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TracerSetting(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tr := otel.Tracer(serviceName) // spanのotel.library.name semantic conventionsに入る値
		ctx := c.Request.Context()     // 現在のリクエストに関連付けられたcontextを取得する（traceのルートとなるコンテキストを生成）
		method := c.Request.Method
		urlPath := c.Request.URL.Path
		ctx, span := tr.Start(ctx, fmt.Sprintf("%s %s", method, urlPath), trace.WithAttributes(
			attribute.String("service.name", serviceName),
		)) // (新しい)spanの開始
		defer span.End() // spanの終了

		c.Request = c.Request.WithContext(ctx) // 現在のHTTPリクエストに新しいコンテキストを設定する（つまり、新しい子コンテキストを作成）
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
