package services

import (
	"context"
	"fmt"
	"ham3/utilities"
	"net/http"
	"time"

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
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logaas_id := c.Param("logaas_id")

	fmt.Printf("ClusterName: %s, ClusterType: %s\n", logaas_id, requestData.ClusterType)

	// Helmの設定
	install, _, chart := utilities.OpenSearchHelmSetting(logaas_id, "install")

	values := map[string]interface{}{
		"replicas": 3,
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "1",
				"memory": "250Mi",
			},
		},
		"extraEnvs": []map[string]interface{}{
			{
				"name":  "OPENSEARCH_INITIAL_ADMIN_PASSWORD",
				"value": "Watchuserstep#3",
			},
		},
		"nameOverride": install.ReleaseName,
	}

	// タイムアウトを10分に設定
	ctxtimeout, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	release, err := install.RunWithContext(ctxtimeout, chart, values)
	// release, err := install.Run(chart, values)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Failed to install chart: %v", err),
		})
		return
	}
	fmt.Printf("Successfully installed chart with release name: %s\n", release.Name)

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
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logaas_id := c.Param("logaas_id")

	fmt.Printf("ClusterName: %s, ClusterType: %s\n", logaas_id, requestData.ClusterType)

	// Helmの設定
	_, uninstall, _ := utilities.OpenSearchHelmSetting(logaas_id, "uninstall")

	_, err := uninstall.Run(logaas_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Failed to uninstall chart: %v", err),
		})
		return
	}
	fmt.Printf("Successfully uninstalled chart with release name: %s\n", logaas_id)

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
