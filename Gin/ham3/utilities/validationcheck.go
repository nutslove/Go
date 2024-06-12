package utilities

import (
	"ham3/services"
)

func CheckLogaasCreateParameters(requestData services.RequestData) (bool, string) {

	var errExist bool
	var errMessage string
	if requestData.ClusterType == "" {
		errExist = true
		errMessage = "cluster-type is required. "
	}

	return errExist, errMessage
}
