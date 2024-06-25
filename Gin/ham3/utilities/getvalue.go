package utilities

import (
	"bytes"
	"fmt"
	"ham3/config"
	"math"
	"os"
	"strconv"
	"text/template"
)

func LogaasGetDefaultValue(requestData *config.LogaasRequestData) {
	// デフォルト値の設定
	requestData.OpenSearchVersion = "2.9.0"
	requestData.OpenSearchDashboardsVersion = "2.9.0"
	requestData.ClusterType = "standard"
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

	type OpensearchData struct {
		Nproc            int
		OpenSearchVer    string
		ExporterAivenVer []string
		ClusterType      string
	}
	var err error
	var values map[string]interface{}

	if requestData.ClusterType == "scalable" {
		opensearchTypeList := []string{"master", "data", "client"}
		var flavor map[string]interface{}
		for _, opensearchType := range opensearchTypeList {
			switch opensearchType {
			case "master":
				flavor = config.Flavors[requestData.MasterFlavor].(map[string]interface{})
			case "data":
				flavor = config.Flavors[requestData.DataFlavor].(map[string]interface{})
			case "client":
				flavor = config.Flavors[requestData.ClientFlavor].(map[string]interface{})
			}
			requests := flavor["requests"].(map[string]string)
			limits := flavor["limits"].(map[string]string)
			requests_cpu := requests["cpu"]
			requests_memory := requests["memory"]
			limits_cpu := limits["cpu"]
			limits_memory := limits["memory"]
			jvm_heap := flavor["jvm_heap"]
			jvm_perm := flavor["jvm_perm"]

			var n_proc int
			if limits_cpu[len(limits_cpu)-1:] == "m" {
				n, err := strconv.Atoi(limits_cpu[:len(limits_cpu)-1])
				if err != nil {
					return nil, err
				}
				n_proc = int(math.Ceil(float64(n) / 1000))
			} else {
				n, err := strconv.Atoi(limits_cpu)
				if err != nil {
					return nil, err
				}
				n_proc = n
			}

			funcMap := template.FuncMap{
				"contains": Contains,
			}
			t, err := template.New(logaas_id).Funcs(funcMap).Parse(config.OpensearchYamlTmpl)
			if err != nil {
				return nil, err
			}

			OpenSearchTmpldata := OpensearchData{
				Nproc:            n_proc,
				OpenSearchVer:    requestData.OpenSearchVersion,
				ExporterAivenVer: config.Exporter.(map[string]interface{})["aiven_ver"],
				ClusterType:      requestData.ClusterType,
			}

			// Templateの結果を格納する変数を定義
			var buf bytes.Buffer
			err = t.Execute(&buf, OpenSearchTmpldata)
			if err != nil {
				return nil, err
			}

			// Templateの結果を文字列に変換
			opensearchYaml := buf.String()

			values = map[string]interface{}{
				"clusterName": logaas_id,
				"nodeGroup":   opensearchType,
				"roles": func() []string {
					var opensearch_type []string
					switch opensearchType {
					case "master":
						opensearch_type = []string{"master"}
					case "data":
						opensearch_type = []string{"ingest", "data"}
					case "client":
						opensearch_type = []string{"ingest"}
					}
					return opensearch_type
				}(),
				"masterService": fmt.Sprintf("%s-master", logaas_id),
				"replicas": func() int {
					var replicas int
					switch opensearchType {
					case "master":
						replicas = 3
					case "data":
						replicas = requestData.ScaleSize
					case "client":
						replicas = 2
					}
					return replicas
				}(),
				"rbac": map[string]interface{}{
					"create": func() bool {
						var rbacCreate bool
						if requestData.OpenSearchVersion == "1.1.0" {
							rbacCreate = false
						} else {
							rbacCreate = true
						}
						return rbacCreate
					}(),
					"serviceAccountName": func() string {
						var saName string
						if requestData.OpenSearchVersion == "1.1.0" {
							saName = "es"
						} else {
							saName = fmt.Sprintf("%s-client", logaas_id)
						}
						return saName
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
					"limits": map[string]string{
						"cpu":    limits_cpu,
						"memory": limits_memory,
					},
					"requests": map[string]string{
						"cpu":    requests_cpu,
						"memory": requests_memory,
					},
				},
				"antiAffinityTopologyKey": "kubernetes.io/hostname",
				"plugins": map[string]interface{}{
					"enabled": true,
					"intallList": func() []string {
						var prometheusExporter []string
						if Contains(config.Exporter["aparo_ver"].([]string), requestData.OpenSearchVersion) {
							prometheusExporter = []string{fmt.Sprintf("https://github.com/aparo/opensearch-prometheus-exporter/releases/download/%s/prometheus-exporter-%s.zip", requestData.OpenSearchVersion, requestData.OpenSearchVersion), "repository-s3"}
						} else if Contains(config.Exporter["aiven_ver"].([]string), requestData.OpenSearchVersion) {
							prometheusExporter = []string{fmt.Sprintf("https://github.com/aiven/prometheus-exporter-plugin-for-opensearch/releases/download/%s.0/prometheus-exporter-%s.0.zip", requestData.OpenSearchVersion, requestData.OpenSearchVersion), "repository-s3"}
						}
						return prometheusExporter
					}(),
				},
				"config": map[string]interface{}{
					"opensearch.yml": opensearchYaml,
				},
				"extraEnvs": []map[string]interface{}{
					{
						"name":  "DISABLE_INSTALL_DEMO_CONFIG",
						"value": "true",
					},
				},
				"extraVolumes": []map[string]interface{}{
					{
						"name": "pem",
						"configMap": map[string]string{
							"name": "opensearch-pem-config",
						},
					},
				},
				"extraVolumeMounts": []map[string]interface{}{
					{
						"name":      "pem",
						"mountPath": "/usr/share/opensearch/config/esnode-key.pem",
						"subPath":   "esnode-key.pem",
					},
					{
						"name":      "pem",
						"mountPath": "/usr/share/opensearch/config/esnode.pem",
						"subPath":   "esnode.pem",
					},
					{
						"name":      "pem",
						"mountPath": "/usr/share/opensearch/config/root-ca.pem",
						"subPath":   "root-ca.pem",
					},
				},
				"securityConfig": map[string]interface{}{
					"config": map[string]interface{}{
						"data": map[string]interface{}{
							"action_groups.yml":  config.ActionGroupsYaml,
							"audit.yml":          auditYaml,
							"config.yml":         configYaml,
							"internal_users.yml": internalUsersYaml,
							"nodes_dn.yml":       nodesDnYaml,
							"roles.yml":          rolesYaml,
							"roles_mapping.yml":  rolesMappingYaml,
							"tenants.yml":        tenantsYaml,
							"whitelist.yml":      whitelistYaml,
						},
					},
				},
			}

			// OpenSearchのバージョンが1.1.0より上の場合、valuesにextraObjectsを追加する
			if requestData.OpenSearchVersion > "1.1.0" {
				extraObjectsValue := []map[string]interface{}{
					{
						"apiVersion": "rbac.authorization.k8s.io/v1",
						"kind":       "RoleBinding",
						"metadata": map[string]string{
							"name":      fmt.Sprintf("scc:nonroot:%s-%s", logaas_id, opensearchType),
							"namespace": "opensearch",
						},
						"roleRef": map[string]string{
							"apiGroup": "rbac.authorization.k8s.io",
							"kind":     "ClusterRole",
							"name":     "system:openshift:scc:nonroot",
						},
						"subjects": []map[string]string{
							{
								"kind":      "ServiceAccount",
								"name":      fmt.Sprintf("%s-%s", logaas_id, opensearchType),
								"namespace": "opensearch",
							},
						},
					},
				}
				AddToMapWithCondition(values, "extraObjects", extraObjectsValue)
			}

			// OpenSearchのバージョンが2.0.0より上の場合、valuesのsecurityConfigフィールドにpathを追加する
			if requestData.OpenSearchVersion > "2.0.0" {
				pathValue := "/usr/share/opensearch/config/opensearch-security"
				AddToMapWithCondition(values["securityConfig"].(map[string]interface{}), "path", pathValue)
			}
		}
	} else if requestData.ClusterType == "standard" {
		opensearchTypeList := []string{"api"}
		for _, opensearchType := range opensearchTypeList {
			values = map[string]interface{}{
				"clusterName": logaas_id,
			}
		}
	} else {
		err := "Invalid cluster type"
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

func AddToMapWithCondition(m map[string]interface{}, key string, value interface{}) {
	m[key] = value
}
