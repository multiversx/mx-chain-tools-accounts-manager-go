package accountsReindex

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-accounts-manager/process"
)

type ExtendedElasticHandler interface {
	process.ElasticClientHandler
	DoBulkRequestDestination(buff *bytes.Buffer, index string) error
	DoScrollRequestAllDocuments(index string, handlerFunc func(responseBytes []byte) error) error
}
