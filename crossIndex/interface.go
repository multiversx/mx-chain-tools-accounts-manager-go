package crossIndex

import (
	"bytes"

	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

// ElasticClientHandler defines what an elastic client should be able to do
type ElasticClientHandler interface {
	PutPolicy(policyName string, policy *bytes.Buffer) error
	PutMapping(targetIndex string, body *bytes.Buffer) error
	CreateIndexWithMapping(index string, mapping *bytes.Buffer) error
	CheckIfIndexExists(index string) (bool, error)
	DoRequest(index, documentID string, buff *bytes.Buffer) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoMultiGet(ids []string, index string) ([]byte, error)
	DoScrollRequestAllDocuments(index string, body []byte, handlerFunc func(responseBytes []byte) error) error
	IsInterfaceNil() bool
}

// AccountsIndexerHandler defines what an accounts' indexer should be able to do
type AccountsIndexerHandler interface {
	GetAccounts(addresses []string, index string) (map[string]*data.AccountInfoWithStakeValues, error)
	IndexAccounts(accounts map[string]*data.AccountInfoWithStakeValues, index string) error
}

// AccountsProcessorHandler defines what an accounts' processor should be able to do
type AccountsProcessorHandler interface {
	GetAllAccountsWithStake() (map[string]*data.AccountInfoWithStakeValues, []string, error)
	PrepareAccountsForReindexing(accountsES, accountsRest map[string]*data.AccountInfoWithStakeValues) map[string]*data.AccountInfoWithStakeValues
	ComputeClonedAccountsIndex() (string, error)
}

// AccountsGetterHandler defines what an accounts' getter should be able to do
type AccountsGetterHandler interface {
	GetLegacyDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error)
	GetValidatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error)
	GetDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error)
}
