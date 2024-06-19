package utilities

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
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

	// KeyStoneサービスクライアントを作成
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
