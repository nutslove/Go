package config

// RequestData is a struct that represents the request data for LOGaaS.
type RequestData struct {
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
		"requests": map[string]interface{}{
			"cpu":    "125m",
			"memory": "640Mi",
		},
		"limits": map[string]interface{}{
			"cpu":    "500m",
			"memory": "1Gi",
		},
		"jvm_heap": "512M",
		"jvm_perm": "128M",
	},
}

// 		m1.small:
// 		requests:
// 			cpu: 250m
// 			memory: 1280Mi
// 		limits:
// 			cpu: 1000m
// 			memory: 2Gi
// 		jvm_heap: 1G
// 		jvm_perm: 256M
// 	m1.medium:
// 		requests:
// 			cpu: 500m
// 			memory: 4352Mi
// 		limits:
// 			cpu: 2000m
// 			memory: 8Gi
// 		jvm_heap: 4G
// 		jvm_perm: 256M
// 	m1.large:
// 		requests:
// 			cpu: 1000m
// 			memory: 8448Mi
// 		limits:
// 			cpu: 4000m
// 			memory: 16Gi
// 		jvm_heap: 8G
// 		jvm_perm: 256M
// 	m1.xlarge:
// 		requests:
// 			cpu: 2000m
// 			memory: 16640Mi
// 		limits:
// 			cpu: 8000m
// 			memory: 32Gi
// 		jvm_heap: 16G
// 		jvm_perm: 256M
// 	d1.tiny:
// 		requests:
// 			cpu: 125m
// 			memory: 1Gi
// 		limits:
// 			cpu: 500m
// 			memory: 1Gi
// 		jvm_heap: 512M
// 		jvm_perm: 128M
// 	d1.small:
// 		requests:
// 			cpu: 250m
// 			memory: 2Gi
// 		limits:
// 			cpu: 1000m
// 			memory: 2Gi
// 		jvm_heap: 1G
// 		jvm_perm: 256M
// 	d1.medium:
// 		requests:
// 			cpu: 500m
// 			memory: 8Gi
// 		limits:
// 			cpu: 2000m
// 			memory: 8Gi
// 		jvm_heap: 4G
// 		jvm_perm: 256M
// 	d1.large:
// 		requests:
// 			cpu: 1000m
// 			memory: 16Gi
// 		limits:
// 			cpu: 4000m
// 			memory: 16Gi
// 		jvm_heap: 8G
// 		jvm_perm: 256M
// 	d1.mlarge:
// 		requests:
// 			cpu: 1000m
// 			memory: 32Gi
// 		limits:
// 			cpu: 4000m
// 			memory: 32Gi
// 		jvm_heap: 16G
// 		jvm_perm: 256M
// 	d1.xlarge:
// 		requests:
// 			cpu: 2000m
// 			memory: 32Gi
// 		limits:
// 			cpu: 8000m
// 			memory: 32Gi
// 		jvm_heap: 16G
// 		jvm_perm: 256M
// `
