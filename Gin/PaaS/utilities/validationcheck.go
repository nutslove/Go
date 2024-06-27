package utilities

import (
	"ham3/config"
)

func CheckLogaasCreateParameters(requestData config.LogaasRequestData) (bool, string) {
	var errExist bool
	var errMessage string
	if requestData.ClusterType != "scalable" || requestData.ClusterType != "standard" {
		errExist = true
		errMessage = "cluster-type must be either 'scalable' or 'standard'."
	}

	return errExist, errMessage
}
