package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":   "HAM3",
		"content": "HAM3 provides services for Logging as a Service (LOGaaS) and Container as a Service (CaaS).",
	})
}
