package utilities

import (
	"ham3/config"
)

func CheckLogaasCreateParameters(requestData config.LogaasRequestData) (bool, string) {
	var errExist bool
	var errMessage string
	if requestData.ClusterType == "" {
		errExist = true
		errMessage = "cluster-type is required. "
	}

	return errExist, errMessage
}
