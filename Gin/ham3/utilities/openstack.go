package utilities

import (
	"errors"
	"fmt"
	"ham3/config"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
)

// 戻り値: ProjectID、Adminなのかどうか、errorメッセージ
func TokenAuth(token string) (string, bool, error) {
	var IsAdmin bool
	var ProjectId string

	// プロバイダーを作成
	provider, err := OpenstackProvider()
	if err != nil {
		errMessage := fmt.Errorf("An error occurred during authentication. err: %v", err)
		return ProjectId, IsAdmin, errMessage
	}

	// KeyStoneサービスクライアントを初期化
	keystoneClient, err := openstack.NewIdentityv3(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OPENSTACK_REGION"),
	})
	if err != nil {
		errMessage := fmt.Errorf("An error occurred during creating keystone client. err: %v", err)
		return ProjectId, IsAdmin, errMessage
	}

	// トークン検証
	tokenValidationResult, err := tokens.Validate(keystoneClient, token)
	if err != nil {
		errMessage := fmt.Errorf("An error occurred during token validation. err: %v", err)
		return ProjectId, IsAdmin, errMessage
	}
	if !tokenValidationResult {
		errMessage := errors.New("Invalid token.")
		return ProjectId, IsAdmin, errMessage
	}

	// トークンの詳細情報を取得
	tokenDetail := tokens.Get(keystoneClient, token)
	tokeninfo, ok := tokenDetail.Body.(map[string]interface{})["token"].(map[string]interface{})
	if !ok {
		errMessage := errors.New("An error occurred while converting to tokenDetail.Body")
		return ProjectId, IsAdmin, errMessage
	}

	ProjectId = tokeninfo["project"].(map[string]interface{})["id"]
	roles := tokeninfo["roles"].([]interface{})

	for _, role := range roles {
		if strings.Contains(role.(map[string]interface{})["name"].(string), "admin") {
			IsAdmin = true
		}
	}

	return ProjectId, IsAdmin, nil
}

func OpenstackProvider() (*gophercloud.ProviderClient, error) {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: os.Getenv("OPENSTACK_AUTH_ENDPOINT"),
		Username:         os.Getenv("OPENSTACK_USERNAME"),
		Password:         os.Getenv("OPENSTACK_PASSWORD"),
		DomainName:       "Default",
		TenantName:       os.Getenv("OPENSTACK_PROJECT_NAME"),
	}

	return openstack.AuthenticatedClient(opts)
}

func CinderClient(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	client, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OPENSTACK_REGION"),
	})
	if err != nil {
		errMessage := fmt.Errorf("An error occurred during creating cinder client. err: %v", err)
		return client, errMessage
	}
	return client, nil
}

func CreateCinderVolume(logaas_id string, requestData config.LogaasRequestData) error {
	provider, err := OpenstackProvider()
	if err != nil {
		errMessage := fmt.Errorf("An error occurred during authentication. err: %v", err)
		return errMessage
	}

	// Cinderサービスクライアントを初期化
	cinderClient, err := CinderClient(provider)
	if err != nil {
		return err
	}

	var volSize int
	var volType string

	if requestData.ClusterType == "scalable" {
		volSize = 2
		volType = "economy-medium"
	} else if requestData.ClusterType == "standard" {
		volSize = requestData.DataDiskSize
		volType = requestData.DiskType
	}

	// Volumeの設定(masterノード用)
	for i := 0; i < 3; i++ {
		masaterCreateOpts := volumes.CreateOpts{
			Description:      fmt.Sprintf("%s-%s-opensearch-pv", requestData.OcpCluster, logaas_id),
			Name:             fmt.Sprintf("%s-%s-master-opensearch-pv-%v", requestData.OcpCluster, logaas_id, i),
			AvailabilityZone: requestData.Zone,
			Size:             volSize,
			VolumeType:       volType,
		}

		volume, err := volumes.Create(cinderClient, masaterCreateOpts).Extract()
		if err != nil {
			errMessage := fmt.Errorf("Failed to create volume: %v", err)
			return errMessage
		}
		fmt.Printf("Created volume: %s with ID: %s\n", fmt.Sprintf("%s-%s-master-opensearch-pv-%v", requestData.OcpCluster, logaas_id, i), volume.ID)
	}

	if requestData.ClusterType == "scalable" {
		// Volumeの設定(dataノード用)
		count := requestData.ScaleSize
		for i := 0; i < count; i++ {
			dataCreateOpts := volumes.CreateOpts{
				Description:      fmt.Sprintf("%s-%s-opensearch-pv", requestData.OcpCluster, logaas_id),
				Name:             fmt.Sprintf("%s-%s-data-opensearch-pv-%v", requestData.OcpCluster, logaas_id, i),
				AvailabilityZone: requestData.Zone,
				Size:             requestData.DataDiskSize,
				VolumeType:       requestData.DiskType,
			}

			volume, err := volumes.Create(cinderClient, dataCreateOpts).Extract()
			if err != nil {
				errMessage := fmt.Errorf("Failed to create volume: %v", err)
				return errMessage
			}
			fmt.Printf("Created volume: %s with ID: %s\n", fmt.Sprintf("%s-%s-data-opensearch-pv-%v", requestData.OcpCluster, logaas_id, i), volume.ID)
		}
	}

	return nil
}

func DeleteCinderVolume(logaas_id string, requestData config.LogaasRequestData) error {
	provider, err := OpenstackProvider()
	if err != nil {
		errMessage := fmt.Errorf("An error occurred during authentication. err: %v", err)
		return errMessage
	}

	// Cinderサービスクライアントを初期化
	cinderClient, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OPENSTACK_REGION"),
	})
	if err != nil {
		errMessage := fmt.Errorf("An error occurred during creating cinder client. err: %v", err)
		return errMessage
	}
	return nil
}
