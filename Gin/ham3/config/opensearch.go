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
