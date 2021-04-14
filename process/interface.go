package process

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// ElasticClientHandler defines what an elastic client should be able do
type ElasticClientHandler interface {
	CloneIndex(index, targetIndex string) (cloned bool, err error)
	WaitYellowStatus() error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoMultiGet(ids []string, index string) ([]byte, error)
	UnsetReadOnly(index string) error
}

// // RestClientHandler defines what a rest client should be able do
type RestClientHandler interface {
	CallGetRestEndPoint(path string, value interface{}) error
	CallPostRestEndPoint(path string, data interface{}, response interface{}) error
}

// AccountsIndexerHandler defines what an accounts indexer should be able do
type AccountsIndexerHandler interface {
	GetAccounts(addresses []string, index string) (map[string]*data.AccountInfo, error)
	IndexAccounts(accounts map[string]*data.AccountInfo, index string) error
}

// AccountsProcessorHandler defines what an accounts processor should be able do
type AccountsProcessorHandler interface {
	GetAllAccountsWithStake() (map[string]*data.AccountInfo, []string, error)
	PrepareAccountsForReindexing(accountsES, accountsRest map[string]*data.AccountInfo) map[string]*data.AccountInfo
	ComputeClonedAccountsIndex() (string, error)
}

type accountsGetterHandler interface {
	getLegacyDelegatorsAccounts() (map[string]string, error)
	getValidatorsAccounts() (map[string]string, error)
	getDelegatorsAccounts() (map[string]string, error)
}

// Cloner defines what a clone should be able to do
type Cloner interface {
	CloneIndex(index, newIndex string) error
}
