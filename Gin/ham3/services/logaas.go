package services

import (
	"context"
	"fmt"
	"ham3/config"
	"ham3/utilities"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
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

func CreateLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset, db *gorm.DB) {
	var requestData config.LogaasRequestData

	// OpenSearchのメタデータ(e.g. cluster type)のデフォルト値を取得
	utilities.LogaasGetDefaultValue(&requestData)

	// OpenSearchのメタデータを実際のリクエスト値に上書き（リクエストに連携されてないパラメータはデフォルト値で設定される）
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logaas_id := c.Param("logaas_id")

	// パラメータのバリデーションチェック
	if errExist, errMessage := utilities.CheckLogaasCreateParameters(requestData); errExist {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMessage})
		return
	}

	fmt.Printf("ClusterName: %s, Cluster Metadata: %s\n", logaas_id, requestData)

	meta := config.Flavors
	flavor := config.Flavors[requestData.MasterFlavor].(map[string]interface{})
	requests := flavor["requests"].(map[string]interface{})
	limits := flavor["limits"].(map[string]interface{})
	requests_cpu := requests["cpu"]
	requests_memory := requests["memory"]
	limits_cpu := limits["cpu"]
	limits_memory := limits["memory"]
	jvm_heap := flavor["jvm_heap"]
	jvm_perm := flavor["jvm_perm"]
	fmt.Println("meta:", meta)
	fmt.Println("flavor:", flavor)
	fmt.Println("requests:", requests)
	fmt.Println("requests_cpu:", requests_cpu)
	fmt.Println("requests_memory:", requests_memory)
	fmt.Println("limits_cpu:", limits_cpu)
	fmt.Println("limits_memory:", limits_memory)
	fmt.Println("jvm_heap:", jvm_heap)
	fmt.Println("jvm_perm:", jvm_perm)

	RbacCreate := true
	ServiceAccountName := fmt.Sprintf("%s-client", logaas_id)
	// OpenSearch 1.1.0の場合はRbacを作成せず、既存のServiceAccount(es)を利用する
	if requestData.OpenSearchVersion == "1.1.0" {
		RbacCreate = false
		ServiceAccountName = "es"
	}

	// Helmの設定
	install, _, chart := utilities.OpenSearchHelmSetting(logaas_id, "install")

	values := map[string]interface{}{
		"clusterName": logaas_id,
		"nodeGroup":   "client",
		"roles": []string{
			"ingest",
		},
		"masterService": fmt.Sprintf("%s-master", logaas_id),
		"replicas":      2,
		"rbac": map[string]interface{}{
			"create":             RbacCreate,
			"serviceAccountName": ServiceAccountName,
		},
		"persistence": map[string]interface{}{
			"enabled":         false,
			"enableInitChown": false,
		},
		"podSecurityContext": map[string]interface{}{
			"runAsUser": 1000,
		},
		"ingress": map[string]interface{}{
			"ingressClassName": "openshift-default",
			"enabled":          true,
			"annotations": map[string]interface{}{
				"route.openshift.io/termination": "edge",
			},
			"hosts": []string{
				fmt.Sprintf("%s-api.es.%s", logaas_id, requestData.BaseDomain),
			},
		},
		// 値要修正
		"opensearchJavaOpts": fmt.Sprintf("-Xms%s -Xmx%s -XX:MaxMetaspaceSize=%s -Dhttp.proxyHost=%s -Dhttp.proxyPort=%s -Dhttps.proxyHost=%s -Dhttps.proxyPort=%s", "1g", "1g", "256m", "proxy.example.com", "8080", "proxy.example.com", "8080"),
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "1",
				"memory": "250Mi",
			},
			"requests": map[string]interface{}{
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
	// ctxtimeout, cancel := context.WithTimeout(ctx, 600*time.Second)
	// defer cancel()

	// release, err := install.RunWithContext(ctxtimeout, chart, values)
	release, err := install.Run(chart, values)
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

func GetLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset, db *gorm.DB) {
	logaas_id := c.Param("logaas_id")

	var requestData config.LogaasRequestData
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

func UpdateLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset, db *gorm.DB) {
	logaas_id := c.Param("logaas_id")

	var requestData config.LogaasRequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("ClusterName: %s, ClusterType: %s\n", logaas_id, requestData.ClusterType)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Updated LOGaaS for %s successfully", logaas_id),
	})
}

func DeleteLogaas(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset, db *gorm.DB) {
	var requestData config.LogaasRequestData
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

func GetLogaases(ctx context.Context, c *gin.Context, clientset *kubernetes.Clientset, db *gorm.DB) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get logaases",
	})
}
