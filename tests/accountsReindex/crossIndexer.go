package accountsReindex

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/process"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type crossIndexer struct {
	extendedClient ExtendedElasticHandler
	count          int
}

var log = logger.GetOrCreate("cross-indexer")

func NewCrossIndexer(sourceDB string, destinationDB string) (*crossIndexer, error) {
	extendedClient, err := NewExtendedElasticClient(sourceDB, destinationDB)
	if err != nil {
		return nil, err
	}

	return &crossIndexer{
		extendedClient: extendedClient,
	}, nil
}

func (ci *crossIndexer) ReindexAccountsIndex() error {
	return ci.extendedClient.DoScrollRequestAllDocuments("accounts", ci.saveAccountsInDestinationDB)
}

func (ci *crossIndexer) saveAccountsInDestinationDB(responseBytes []byte) error {
	ci.count++
	log.Info("Indexing accounts", "bulk", ci.count)

	accounts, err := getAllAccounts(responseBytes)
	if err != nil {
		return err
	}

	return ci.indexAllAccounts(accounts)
}

func (ci *crossIndexer) indexAllAccounts(mapAllAccounts map[string]*data.AccountInfoWithStakeValues) error {
	acIndexer, err := process.NewAccountsIndexer(ci.extendedClient)
	if err != nil {
		return err
	}

	return acIndexer.IndexAccounts(mapAllAccounts, "accounts-000001")
}

func getAllAccounts(responseBytes []byte) (map[string]*data.AccountInfoWithStakeValues, error) {
	accountsResponse := &AllAccountsResponse{}
	err := json.Unmarshal(responseBytes, &accountsResponse)
	if err != nil {
		return nil, err
	}

	accts := make(map[string]*data.AccountInfoWithStakeValues)
	for _, acct := range accountsResponse.Hits.Hits {

		acc := data.AccountInfoWithStakeValues{}

		acc = acct.Account

		accts[acct.ID] = &acc
	}

	return accts, nil
}
