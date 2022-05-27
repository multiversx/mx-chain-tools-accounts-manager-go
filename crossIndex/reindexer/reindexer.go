package reindexer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-accounts-manager/core"
	"github.com/ElrondNetwork/elrond-accounts-manager/crossIndex"
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/process/accountsIndexer"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/check"
)

type reindexer struct {
	sourceIndexer       crossIndex.ElasticClientHandler
	destinationClients  []crossIndex.ElasticClientHandler
	count               int
	pathToIndicesConfig string
}

var log = logger.GetOrCreate("reindexer")

// New returns a new instance of reindexer
func New(
	sourceIndexer crossIndex.ElasticClientHandler,
	destinationIndexer []crossIndex.ElasticClientHandler,
	pathToIndicesConfig string) (*reindexer, error,
) {
	if check.IfNil(sourceIndexer) {
		return nil, fmt.Errorf("%w for sourceIndexer", crossIndex.ErrNilElasticClient)
	}
	if pathToIndicesConfig == "" {
		return nil, errors.New("empty path to the indices config folder")
	}
	for idx, dstClient := range destinationIndexer {
		if check.IfNil(dstClient) {
			return nil, fmt.Errorf("%w for destinationIndexer, index %d", crossIndex.ErrNilElasticClient, idx)
		}
	}

	return &reindexer{
		sourceIndexer:       sourceIndexer,
		destinationClients:  destinationIndexer,
		pathToIndicesConfig: pathToIndicesConfig,
	}, nil
}

// ReindexAccounts will reindex all accounts from source indexer to destination indexer
func (r *reindexer) ReindexAccounts(sourceIndex string, destinationIndex string, restAccounts map[string]*data.AccountInfoWithStakeValues) error {
	log.Info("Create a new index with mapping")

	template, policy, err := readTemplateAndPolicy(r.pathToIndicesConfig)
	if err != nil {
		return err
	}

	templateBytes := template.Bytes()
	policyBytes := policy.Bytes()

	for _, dstClient := range r.destinationClients {
		err = dstClient.CreateIndexWithMapping(destinationIndex, bytes.NewBuffer(templateBytes))
		if err != nil {
			return err
		}

		err = dstClient.PutPolicy(crossIndex.AccountsPolicyName, bytes.NewBuffer(policyBytes))
		if err != nil {
			return err
		}
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
	for _, dstClient := range r.destinationClients {
		acIndexer, err := accountsIndexer.NewAccountsIndexer(dstClient)
		if err != nil {
			return err
		}

		err = acIndexer.IndexAccounts(mapAllAccounts, destinationIndex)
		if err != nil {
			return err
		}
	}
	return nil
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
