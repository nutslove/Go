package services

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateLogaas(c *gin.Context) {
	logaas_id := c.Param("logaas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Create logaas %s", logaas_id),
	})
}

func GetLogaas(c *gin.Context) {
	logaas_id := c.Param("logaas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Get logaas %s", logaas_id),
	})
}

func DeleteLogaas(c *gin.Context) {
	logaas_id := c.Param("logaas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Delete logaas %s", logaas_id),
	})
}

func GetLogaases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get logaases",
	})
}
