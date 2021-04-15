package process

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-accounts-manager/data"
)

// ElasticClientHandler defines what an elastic client should be able do
type ElasticClientHandler interface {
	CloneIndex(index, targetIndex string) (cloned bool, err error)
	PutMapping(targetIndex string, body *bytes.Buffer) error
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
	GetAccounts(addresses []string, index string) (map[string]*data.AccountInfoWithStakeValues, error)
	IndexAccounts(accounts map[string]*data.AccountInfoWithStakeValues, index string) error
}

// AccountsProcessorHandler defines what an accounts processor should be able do
type AccountsProcessorHandler interface {
	GetAllAccountsWithStake() (map[string]*data.AccountInfoWithStakeValues, []string, error)
	PrepareAccountsForReindexing(accountsES, accountsRest map[string]*data.AccountInfoWithStakeValues) map[string]*data.AccountInfoWithStakeValues
	ComputeClonedAccountsIndex() (string, error)
}

type accountsGetterHandler interface {
	getLegacyDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error)
	getValidatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error)
	getDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error)
}

// Cloner defines what a clone should be able to do
type Cloner interface {
	CloneIndex(index, newIndex string, body *bytes.Buffer) error
}
