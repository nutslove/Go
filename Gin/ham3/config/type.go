package config

// RequestData is a struct that represents the request data for LOGaaS.
type RequestData struct {
	ClusterType                 string `json:"cluster-type"`
	OpenSearchVersion           string `json:"opensearch-version"`
	OpenSearchDashboardsVersion string `json:"opensearch-dashboards-version"`
	BaseDomain                  string `json:"base-domain"`
}
