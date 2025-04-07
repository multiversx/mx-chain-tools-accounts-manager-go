package mocks

import (
	"bytes"
)

// ElasticClientStub -
type ElasticClientStub struct {
	DoScrollRequestAllDocumentsCalled func(index string, body []byte, handlerFunc func(responseBytes []byte) error) error
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
func (e *ElasticClientStub) DoScrollRequestAllDocuments(index string, body []byte, handlerFunc func(responseBytes []byte) error) error {
	if e.DoScrollRequestAllDocumentsCalled != nil {
		return e.DoScrollRequestAllDocumentsCalled(index, body, handlerFunc)
	}

	return nil
}

// IsInterfaceNil -
func (e *ElasticClientStub) IsInterfaceNil() bool {
	return e == nil
}
