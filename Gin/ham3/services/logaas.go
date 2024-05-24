package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
)

func CreateLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	logaas_id := c.Param("logaas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Create logaas %s", logaas_id),
	})
}

func GetLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	logaas_id := c.Param("logaas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Get logaas %s", logaas_id),
	})
}

func UpdateLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	logaas_id := c.Param("logaas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Update logaas %s", logaas_id),
	})
}

func DeleteLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	logaas_id := c.Param("logaas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Delete logaas %s", logaas_id),
	})
}

func GetLogaases(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get logaases",
	})
}
