package utilities

import (
	"fmt"
	"ham3/config"
	"os"
	"strconv"
	"math"
)

func LogaasGetDefaultValue(requestData *config.LogaasRequestData) {
	// デフォルト値の設定
	requestData.OpenSearchVersion = "2.9.0"
	requestData.OpenSearchDashboardsVersion = "2.9.0"
	// requestData.ClusterType = "standard"
	requestData.ScaleSize = 3
	requestData.BaseDomain = os.Getenv("BASE_DOMAIN")
	requestData.K8sName = os.Getenv("KUBE_NAME")
	requestData.MasterFlavor = "m1.small"
	requestData.ClientFlavor = "m1.small"
	requestData.DataFlavor = "d1.medium"
	requestData.GuiFlavor = "m1.small"
	requestData.DataDiskSize = 8
	requestData.DiskType = "economy-medium"
	requestData.Site = os.Getenv("SITE")
	requestData.Zone = "az-a"
	requestData.OcpCluster = os.Getenv("OCP_CLUSTER")
}

func OpensearchGetHelmValue(logaas_id string, requestData config.LogaasRequestData) (map[string]interface{}, error) {

	var err error
	var values map[string]interface{}


	if requestData.ClusterType == "scalable" {
		opensearchTypeList := []string{"master", "data", "client"}
		for _, opensearchType := range opensearchTypeList {
			switch opensearchType {
			case "master":
				flavor := config.Flavors[requestData.MasterFlavor].(map[string]interface{})
			case "data":
				flavor := config.Flavors[requestData.DataFlavor].(map[string]interface{})
			case "client":
				flavor := config.Flavors[requestData.ClientFlavor].(map[string]interface{})
			}
			requests := flavor["requests"].(map[string]interface{})
			limits := flavor["limits"].(map[string]interface{})
			requests_cpu := requests["cpu"]
			requests_memory := requests["memory"]
			limits_cpu := limits["cpu"]
			limits_memory := limits["memory"]
			jvm_heap := flavor["jvm_heap"]
			jvm_perm := flavor["jvm_perm"]

			n_proc, err := strconv.Atoi(limits_cpu[len(limits_cpu)-1:])
			if err != nil {
				return nil, err
			}

			if limits_cpu[len(limits_cpu)-1:] == "m" {
				n_proc = math.Ceil(float64(n_proc) / 1000)
			}

			values := map[string]interface{}{
				"clusterName": logaas_id,
				"nodeGroup":   opensearchType,
				"roles": func() []string {
					switch opensearchType {
					case "master":
						return []string{"master"}
					case "data":
						return []string{"ingest", "data"}
					case "client":
						return []string{"ingest"}
					}
				}(),
				"masterService": fmt.Sprintf("%s-master", logaas_id),
				"replicas": func() int {
					switch opensearchType {
					case "master":
						return 3
					case "data":
						return requestData.ScaleSize
					case "client":
						return 2
					}
				}(),
				"rbac": map[string]interface{}{
					"create": func() bool {
						if requestData.OpenSearchVersion == "1.1.0" {
							return false
						} else {
							return true
						}
					}(),
					"serviceAccountName": func() string {
						if requestData.OpenSearchVersion == "1.1.0" {
							return "es"
						} else {
							return fmt.Sprintf("%s-client", logaas_id)
						}
					}(),
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
				"opensearchJavaOpts": fmt.Sprintf("-Xms%s -Xmx%s -XX:MaxMetaspaceSize=%s -Dhttp.proxyHost=%s -Dhttp.proxyPort=%s -Dhttps.proxyHost=%s -Dhttps.proxyPort=%s", jvm_heap, jvm_heap, jvm_perm, config.HttpProxyUrl, config.HttpProxyPort, config.HttpProxyUrl, config.HttpProxyPort),
				"resources": map[string]interface{}{
					"limits": map[string]interface{}{
						"cpu":    limits_cpu,
						"memory": limits_memory,
					},
					"requests": map[string]interface{}{
						"cpu":    requests_cpu,
						"memory": requests_memory,
					},
				},
				"antiAffinityTopologyKey": "kubernetes.io/hostname",
				"plugins": map[string]interface{}{
					"enabled": true,
					"intallList": func() []string {
						if Contains(config.Exporter["aparo_ver"].([]string), requestData.OpenSearchVersion) {
							return []string{fmt.Sprintf("https://github.com/aparo/opensearch-prometheus-exporter/releases/download/%s/prometheus-exporter-%s.zip", requestData.OpenSearchVersion, requestData.OpenSearchVersion), "repository-s3"}
						} else if Contains(config.Exporter["aiven_ver"].([]string), requestData.OpenSearchVersion) {
							return []string{fmt.Sprintf("https://github.com/aiven/prometheus-exporter-plugin-for-opensearch/releases/download/%s.0/prometheus-exporter-%s.0.zip", requestData.OpenSearchVersion, requestData.OpenSearchVersion), "repository-s3"}
						}
					}(),
				},
				"config": map[string]interface{}{
					"opensearch.yml": map[string]interface{}{
						"cluster.name": "opensearch-cluster",
						"netwrok.host": "0.0.0.0",
						"node": map[string]interface{}{
							"processors": n_proc,
						},
					},
				"extraEnvs": []map[string]interface{}{
					{
						"name":  "OPENSEARCH_INITIAL_ADMIN_PASSWORD",
						"value": "Watchuserstep#3",
					},
				},
				// "nameOverride": logaas_id,
			}
		}
	} else if requestData.ClusterType == "standard" {
		opensearchTypeList := []string{"api"}
		for _, opensearchType := range opensearchTypeList {
			values := map[string]interface{}{
				"clusterName": logaas_id,
			}
		}
	} else {
		err = "Invalid cluster type"
	}

	return values, err
}

func Contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
