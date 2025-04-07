package reindexer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/crossIndex"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/process/accountsIndexer"
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
func (r *reindexer) ReindexAccounts(sourceIndex string, destinationIndex string, restAccounts *data.AccountsData) error {
	log.Info("Create a new index with mapping")

	template, _, err := readTemplateAndPolicyForAccountsIndex(r.pathToIndicesConfig)
	if err != nil {
		return err
	}

	templateBytes := template.Bytes()

	for _, dstClient := range r.destinationClients {
		err = dstClient.CreateIndexWithMapping(destinationIndex, bytes.NewBuffer(templateBytes))
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

		mergedAccounts := core.MergeElasticAndRestAccounts(esAccounts, restAccounts.AccountsWithStake)

		return r.indexAllAccounts(mergedAccounts, destinationIndex)
	}

	err = r.sourceIndexer.DoScrollRequestAllDocuments(sourceIndex, crossIndex.GetAll().Bytes(), saverFunc)
	if err != nil {
		return err
	}

	err = r.checkAndCreateValuesIndex()
	if err != nil {
		return err
	}

	return r.indexExtraInformation(restAccounts)
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

func (r *reindexer) indexExtraInformation(accountsData *data.AccountsData) error {
	for _, dstClient := range r.destinationClients {
		err := indexEnergyBlockInfo(accountsData.EnergyBlockInfo, accountsData.Epoch, dstClient)
		if err != nil {
			return err
		}
	}

	return nil
}

func indexEnergyBlockInfo(energyBlockInfo *data.BlockInfo, epoch uint32, esClient crossIndex.ElasticClientHandler) error {
	log.Info(fmt.Sprintf("Indexing extra information in `%s` index...", valuesIndex))

	id := fmt.Sprintf("energy-snapshot-%d", epoch)
	keyValueObj := &data.KeyValueObj{
		Key:   "blockHash",
		Value: energyBlockInfo.Hash,
	}

	keyValueObjBytes, err := json.Marshal(keyValueObj)
	if err != nil {
		return err
	}

	return esClient.DoRequest(valuesIndex, id, bytes.NewBuffer(keyValueObjBytes))
}

func (r *reindexer) checkAndCreateValuesIndex() error {
	template, err := readTemplateForIndex(r.pathToIndicesConfig, valuesIndex)
	templateBytes := template.Bytes()

	for _, dstClient := range r.destinationClients {
		exists, errC := dstClient.CheckIfIndexExists(valuesIndex)
		if errC != nil {
			return errC
		}
		if exists {
			continue
		}

		err = dstClient.CreateIndexWithMapping(valuesIndex, bytes.NewBuffer(templateBytes))
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
