package utilities

import (
	"ham3/config"
	"os"
)

var flavor = `

`

func OpenSearchGetDefaultValue(requestData *config.RequestData) {
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
