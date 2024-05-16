package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tr := otel.Tracer("ham3")
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
