package mocks

import (
	"bytes"
)

// ElasticClientStub -
type ElasticClientStub struct {
}

// PutMapping -
func (e *ElasticClientStub) PutMapping(_ string, _ *bytes.Buffer) error {
	panic("implement me")
}

// DoBulkRequest -
func (e *ElasticClientStub) DoBulkRequest(_ *bytes.Buffer, _ string) error {
	panic("implement me")
}

// DoMultiGet -
func (e *ElasticClientStub) DoMultiGet(_ []string, _ string) ([]byte, error) {
	panic("implement me")
}

// DoScrollRequestAllDocuments -
func (e *ElasticClientStub) DoScrollRequestAllDocuments(_ string, _ []byte, _ func(responseBytes []byte) error) error {
	panic("implement me")
}

func (e *ElasticClientStub) IsInterfaceNil() bool {
	return e == nil
}
