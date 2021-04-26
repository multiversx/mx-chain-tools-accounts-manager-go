package core

import "github.com/ElrondNetwork/elrond-accounts-manager/data"

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
