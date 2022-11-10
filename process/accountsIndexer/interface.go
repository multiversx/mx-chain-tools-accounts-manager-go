package accountsIndexer

import "bytes"

// ElasticClientHandler defines what an elastic client should be able to do
type ElasticClientHandler interface {
	PutMapping(targetIndex string, body *bytes.Buffer) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoMultiGet(ids []string, index string) ([]byte, error)
	DoScrollRequestAllDocuments(index string, body []byte, handlerFunc func(responseBytes []byte) error) error
	IsInterfaceNil() bool
}
