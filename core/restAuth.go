package core

import (
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/config"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

// ShouldUseBasicAuthentication returns true if the credentials aren't empty
func ShouldUseBasicAuthentication(authData data.RestApiAuthenticationData) bool {
	return len(authData.Username) > 0 && len(authData.Password) > 0
}

// GetEmptyApiCredentials returns a new object containing empty credentials, so requests won't include authentication
func GetEmptyApiCredentials() data.RestApiAuthenticationData {
	return data.RestApiAuthenticationData{
		Username: "",
		Password: "",
	}
}

// FetchAuthenticationData will extract the rest api authentication data from the configuration
func FetchAuthenticationData(apiConfig config.APIConfig) data.RestApiAuthenticationData {
	return data.RestApiAuthenticationData{
		Username: apiConfig.Username,
		Password: apiConfig.Password,
	}
}
