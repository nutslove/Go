package config

import (
	"os"
)

var (
	HttpProxyUrl  = os.Getenv("HTTP_PROXY_URL")
	HttpProxyPort = os.Getenv("HTTP_PROXY_PORT")
)

// RequestData is a struct that represents the request data for LOGaaS.
type LogaasRequestData struct {
	ClusterType                 string `json:"cluster-type"`
	OpenSearchVersion           string `json:"opensearch-version"`
	OpenSearchDashboardsVersion string `json:"opensearch-dashboards-version"`
	ScaleSize                   int    `json:"scale-size"`
	BaseDomain                  string `json:"base-domain"`
	K8sName                     string `json:"k8s-name"`
	MasterFlavor                string `json:"master-flavor"`
	ClientFlavor                string `json:"client-flavor"`
	DataFlavor                  string `json:"data-flavor"`
	GuiFlavor                   string `json:"gui-flavor"`
	DataDiskSize                int    `json:"data-disk-size"`
	DiskType                    string `json:"disk-type-ham3"`
	Site                        string `json:"site"`
	Zone                        string `json:"zone"`
	OcpCluster                  string `json:"ocp-cluster"`
}

var Flavors = map[string]interface{}{
	"m1.tiny": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "125m",
			"memory": "640Mi",
		},
		"limits": map[string]string{
			"cpu":    "500m",
			"memory": "1Gi",
		},
		"jvm_heap": "512M",
		"jvm_perm": "128M",
	},
	"m1.small": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "250m",
			"memory": "1280Mi",
		},
		"limits": map[string]string{
			"cpu":    "1000m",
			"memory": "2Gi",
		},
		"jvm_heap": "1G",
		"jvm_perm": "256M",
	},
	"m1.medium": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "500m",
			"memory": "4352Mi",
		},
		"limits": map[string]string{
			"cpu":    "2000m",
			"memory": "8Gi",
		},
		"jvm_heap": "4G",
		"jvm_perm": "256M",
	},
	"m1.large": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "1000m",
			"memory": "8448Mi",
		},
		"limits": map[string]string{
			"cpu":    "4000m",
			"memory": "16Gi",
		},
		"jvm_heap": "8G",
		"jvm_perm": "256M",
	},
	"m1.xlarge": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "2000m",
			"memory": "16640Mi",
		},
		"limits": map[string]string{
			"cpu":    "8000m",
			"memory": "32Gi",
		},
		"jvm_heap": "16G",
		"jvm_perm": "256M",
	},
	"d1.tiny": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "125m",
			"memory": "1Gi",
		},
		"limits": map[string]string{
			"cpu":    "500m",
			"memory": "1Gi",
		},
		"jvm_heap": "512M",
		"jvm_perm": "128M",
	},
	"d1.small": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "250m",
			"memory": "2Gi",
		},
		"limits": map[string]string{
			"cpu":    "1000m",
			"memory": "2Gi",
		},
		"jvm_heap": "1G",
		"jvm_perm": "256M",
	},
	"d1.medium": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "500m",
			"memory": "8Gi",
		},
		"limits": map[string]string{
			"cpu":    "2000m",
			"memory": "8Gi",
		},
		"jvm_heap": "4G",
		"jvm_perm": "256M",
	},
	"d1.large": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "1000m",
			"memory": "16Gi",
		},
		"limits": map[string]string{
			"cpu":    "4000m",
			"memory": "16Gi",
		},
		"jvm_heap": "8G",
		"jvm_perm": "256M",
	},
	"d1.mlarge": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "1000m",
			"memory": "32Gi",
		},
		"limits": map[string]string{
			"cpu":    "4000m",
			"memory": "32Gi",
		},
		"jvm_heap": "16G",
		"jvm_perm": "256M",
	},
	"d1.xlarge": map[string]interface{}{
		"requests": map[string]string{
			"cpu":    "2000m",
			"memory": "32Gi",
		},
		"limits": map[string]string{
			"cpu":    "8000m",
			"memory": "32Gi",
		},
		"jvm_heap": "16G",
		"jvm_perm": "256M",
	},
}

var Exporter = map[string]interface{}{
	"aparo_ver": []string{"1.1.0", "1.2.4"},
	"aiven_ver": []string{"1.3.15", "2.3.0", "2.5.0", "2.7.0", "2.9.0", "2.11.1"},
}

var HelmChartVersions = map[string]interface{}{
	"opensearch": map[string]string{
		// "opensearch version": "helm chart version"
		"1.1.0":  "opensearch-1.5.3",
		"1.2.4":  "opensearch-1.8.3",
		"1.3.15": "opensearch-1.25.0",
		"2.3.0":  "opensearch-2.7.0",
		"2.5.0":  "opensearch-2.10.0",
		"2.7.0":  "opensearch-2.12.1",
		"2.9.0":  "opensearch-2.14.1",
		"2.11.1": "opensearch-2.17.3",
	},
	"dashboards": map[string]string{
		// "opensearch-dashboards version": "helm chart version"
		"1.1.0":  "opensearch-dashboards-1.1.0",
		"1.2.0":  "opensearch-dashboards-1.2.2",
		"1.3.15": "opensearch-dashboards-1.18.0",
		"2.3.0":  "opensearch-dashboards-2.5.3",
		"2.5.0":  "opensearch-dashboards-2.8.0",
		"2.7.0":  "opensearch-dashboards-2.10.0",
		"2.9.0":  "opensearch-dashboards-2.12.0",
		"2.11.1": "opensearch-dashboards-2.15.1",
	},
}

var OpensearchYamlTmpl = `
cluster.name: opensearch-cluster
netwrok.host: 0.0.0.0
node:
	processors: {{ .Nproc }}
{{- if Contains .ExporterAivenVer .OpenSearchVer}}
prometheus.metric_name.prefix: "es_"
{{- end}}
{{- if eq .ClusterType "standard"}}
plugins.security.ssl.http.enabled: false // standardのみ存在
{{- end}} 
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

var ActionGroupsYaml = `
---
_meta:
  type: "actiongroups"
	config_version: 2
`
