package reindexer

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-accounts-manager/core"
	"github.com/ElrondNetwork/elrond-accounts-manager/crossIndex"
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/process/accountsIndexer"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/check"
)

type reindexer struct {
	sourceIndexer      crossIndex.ElasticClientHandler
	destinationIndexer crossIndex.ElasticClientHandler
	count              int
}

var log = logger.GetOrCreate("reindexer")

// New returns a new instance of reindexer
func New(sourceIndexer crossIndex.ElasticClientHandler, destinationIndexer crossIndex.ElasticClientHandler) (*reindexer, error) {
	if check.IfNil(sourceIndexer) {
		return nil, fmt.Errorf("%w for sourceIndexer", crossIndex.ErrNilElasticClient)
	}
	if check.IfNil(destinationIndexer) {
		return nil, fmt.Errorf("%w for destinationIndexer", crossIndex.ErrNilElasticClient)
	}

	return &reindexer{
		sourceIndexer:      sourceIndexer,
		destinationIndexer: destinationIndexer,
	}, nil
}

// ReindexAccounts will reindex all accounts from source indexer to destination indexer
func (r *reindexer) ReindexAccounts(sourceIndex string, destinationIndex string, restAccounts map[string]*data.AccountInfoWithStakeValues) error {
	log.Info("Create a new index with mapping")
	err := r.destinationIndexer.CreateIndexWithMapping(destinationIndex, crossIndex.AccountsTemplate.ToBuffer())
	if err != nil {
		return err
	}

	saverFunc := func(responseBytes []byte) error {
		r.count++
		log.Info("indexing accounts", "bulk", r.count)

		esAccounts, errG := getAllAccounts(responseBytes)
		if errG != nil {
			return errG
		}

		mergedAccounts := core.MergeElasticAndRestAccounts(esAccounts, restAccounts)

		return r.indexAllAccounts(mergedAccounts, destinationIndex)
	}

	return r.sourceIndexer.DoScrollRequestAllDocuments(sourceIndex, crossIndex.GetAll().Bytes(), saverFunc)
}

func (r *reindexer) indexAllAccounts(mapAllAccounts map[string]*data.AccountInfoWithStakeValues, destinationIndex string) error {
	acIndexer, err := accountsIndexer.NewAccountsIndexer(r.destinationIndexer)
	if err != nil {
		return err
	}

	return acIndexer.IndexAccounts(mapAllAccounts, destinationIndex)
}

func getAllAccounts(responseBytes []byte) (map[string]*data.AccountInfoWithStakeValues, error) {
	accountsResponse := &crossIndex.AllAccountsResponse{}
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

// IsInterfaceNil returns true if the value under the interface is nil
func (r *reindexer) IsInterfaceNil() bool {
	return r == nil
}
