package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/client-go/kubernetes"
)

var (
	LOGaasCreateCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "logaas_create_total",
		Help: "The total number of LOGaaS created",
	},
		[]string{
			"cluster_name",
			"cluster_type",
		},
	)

	LOGaasDeleteCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "logaas_delete_total",
		Help: "The total number of LOGaaS deleted",
	},
		[]string{
			"cluster_name",
			"cluster_type",
		},
	)
)

func IncreaseLOGaaSCreateCounter(clusterName, clusterType string) {
	LOGaasCreateCounter.WithLabelValues(clusterName, clusterType).Inc()
}

func IncreaseLOGaaSDeleteCounter(clusterName, clusterType string) {
	LOGaasDeleteCounter.WithLabelValues(clusterName, clusterType).Inc()
}

type RequestData struct {
	ClusterType string `json:"cluster-type"`
}

func CreateLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	logaas_id := c.Param("logaas_id")

	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("ClusterName: %s, ClusterType: %s\n", logaas_id, requestData.ClusterType)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Created LOGaaS for %s successfully", logaas_id),
	})
	IncreaseLOGaaSCreateCounter(logaas_id, requestData.ClusterType)
}

func GetLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	logaas_id := c.Param("logaas_id")

	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("ClusterName: %s, ClusterType: %s\n", logaas_id, requestData.ClusterType)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Get LOGaaS info about %s", logaas_id),
	})
}

func UpdateLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	logaas_id := c.Param("logaas_id")

	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("ClusterName: %s, ClusterType: %s\n", logaas_id, requestData.ClusterType)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Updated LOGaaS for %s successfully", logaas_id),
	})
}

func DeleteLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	logaas_id := c.Param("logaas_id")

	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("ClusterName: %s, ClusterType: %s\n", logaas_id, requestData.ClusterType)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Delete LOGaaS for %s successfully", logaas_id),
	})
	IncreaseLOGaaSDeleteCounter(logaas_id, requestData.ClusterType)
}

func GetLogaases(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get logaases",
	})
}
