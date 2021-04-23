package accountsIndexer

import "bytes"

// ElasticClientHandler defines what an elastic client should be able do
type ElasticClientHandler interface {
	CloneIndex(index, targetIndex string) (cloned bool, err error)
	PutMapping(targetIndex string, body *bytes.Buffer) error
	WaitYellowStatus() error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoMultiGet(ids []string, index string) ([]byte, error)
	DoScrollRequestAllDocuments(index string, body []byte, handlerFunc func(responseBytes []byte) error) error
	UnsetReadOnly(index string) error
	IsInterfaceNil() bool
}
