package utilities

import (
	"fmt"
	"ham3/config"
	"math"
	"os"
	"strconv"
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

			opensearchYaml := GetOpensearchYamlTmpl()

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
							"action_groups.yml":  actionGroupsYaml,
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

func GetOpensearchYamlTmpl() string {
	opensearchYamlTmpl := `
	cluster.name: opensearch-cluster
	netwrok.host: 0.0.0.0
	node:
		processors: {{ .Nproc }}
{{if .}}
	prometheus.metric_name.prefix: "es_"
{{end}}
{{if .}}
	plugins.security.ssl.http.enabled: false // standardのみ存在
{{end}} 
	plugins:
		security:
			ssl:
				transport:
					pemcert_filepath: esnode.pem
					pemkey_filepath: esnode-key.pem
					pemtrustedcas_filepath: root-ca.pem
					enforce_hostname_verification: false
				http:
					enabled: false
			allow_unsafe_democertificates: true
			allow_default_init_securityindex: true
			authcz:
				admin_dn:
					- CN=kirk,OU=client,O=client,L=test,C=de
			audit.type: internal_opensearch
			enable_snapshot_restore_privilege: true
			check_snapshot_restore_write_privileges: true
			restapi:
				roles_enabled: ["all_access", "security_rest_api_access"]
			system_indices:
				enabled: true
				indices:
					[
						".opendistro-alerting-config",
						".opendistro-alerting-alert*",
						".opendistro-anomaly-results*",
						".opendistro-anomaly-detector*",
						".opendistro-anomaly-checkpoints",
						".opendistro-anomaly-detection-state",
						".opendistro-reports-*",
						".opendistro-notifications-*",
						".opendistro-nootbooks",
						".opendistro-asynchronous-search-response*",
					]
`
	return opensearchYamlTmpl
}
