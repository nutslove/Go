package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const TokenKey = "X-Auth-Token"

func CheckTokenExists() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := c.GetHeader(TokenKey)
		if tokenValue == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
			})
			return
		}
		// Tokenが存在する場合は次のミドルウェアを呼び出す
		c.Next()
	}
}
