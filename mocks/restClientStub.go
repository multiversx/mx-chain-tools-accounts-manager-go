package mocks

import "github.com/ElrondNetwork/elrond-accounts-manager/data"

type RestClientStub struct {
}

func (r RestClientStub) CallGetRestEndPoint(_ string, _ interface{}, _ data.RestApiAuthenticationData) error {
	panic("implement me")
}

func (r RestClientStub) CallPostRestEndPoint(_ string, _ interface{}, _ interface{}, _ data.RestApiAuthenticationData) error {
	panic("implement me")
}
