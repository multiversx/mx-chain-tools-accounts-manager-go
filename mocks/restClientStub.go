package mocks

import "github.com/multiversx/mx-chain-tools-accounts-manager-go/data"

type RestClientStub struct {
	CallPostRestEndPointCalled func(path string, data interface{}, response interface{}, authenticationData data.RestApiAuthenticationData) error
}

func (r RestClientStub) CallGetRestEndPoint(_ string, _ interface{}, _ data.RestApiAuthenticationData) error {
	panic("implement me")
}

func (r RestClientStub) CallPostRestEndPoint(path string, data interface{}, response interface{}, authenticationData data.RestApiAuthenticationData) error {
	if r.CallPostRestEndPointCalled != nil {
		return r.CallPostRestEndPointCalled(path, data, response, authenticationData)
	}

	return nil
}
