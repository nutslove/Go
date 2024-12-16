package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const tenantId = "Tenant-id"

func TenantidCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(tenantId) == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Tenant-id header missing!",
			})
			return
		}

		c.Next()
	}
}
