package mocks

type RestClientStub struct {
}

func (r RestClientStub) CallGetRestEndPoint(path string, value interface{}) error {
	panic("implement me")
}

func (r RestClientStub) CallPostRestEndPoint(path string, data interface{}, response interface{}) error {
	panic("implement me")
}
