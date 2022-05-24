package process

import (
	"errors"
	"github.com/ElrondNetwork/elrond-accounts-manager/config"
	"github.com/ElrondNetwork/elrond-accounts-manager/core"
	"github.com/ElrondNetwork/elrond-accounts-manager/crossIndex"
	"github.com/ElrondNetwork/elrond-accounts-manager/crossIndex/cloner"
	"github.com/ElrondNetwork/elrond-accounts-manager/crossIndex/reindexer"
	"github.com/ElrondNetwork/elrond-accounts-manager/elasticClient"
	"github.com/ElrondNetwork/elrond-accounts-manager/process/accountsIndexer"
	"github.com/ElrondNetwork/elrond-accounts-manager/restClient"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
)

// CreateDataProcessor will create a new instance of a data processor
func CreateDataProcessor(cfg *config.Config, indexType, indicesConfigPath string) (DataProcessor, error) {
	if indexType == "clone" {
		return getClonerDataProcessor(cfg)
	}

	return getReindexerDataProcessor(cfg, indicesConfigPath)
}

func getClonerDataProcessor(cfg *config.Config) (DataProcessor, error) {
	esClient, err := elasticClient.NewElasticClient(cfg.Cloner.ElasticSearchClient)
	if err != nil {
		return nil, err
	}

	rClient, err := restClient.NewRestClient(cfg.APIConfig.URL)
	if err != nil {
		return nil, err
	}

	pubKeyConverter, err := factory.NewPubkeyConverter(cfg.AddressPubkeyConverter)
	if err != nil {
		return nil, err
	}

	authenticationData := core.FetchAuthenticationData(cfg.APIConfig)
	acctGetter, err := NewAccountsGetter(rClient, cfg.GeneralConfig.DelegationLegacyContractAddress, pubKeyConverter, authenticationData, cfg.GeneralConfig.LKMEXStakingContractAddress)

	acctsProcessor, err := NewAccountsProcessor(rClient, acctGetter)
	if err != nil {
		return nil, err
	}

	acctsIndexer, err := accountsIndexer.NewAccountsIndexer(esClient)
	if err != nil {
		return nil, err
	}

	indexCloner, err := cloner.New(esClient)
	if err != nil {
		return nil, err
	}

	return NewClonerDataProcessor(acctsIndexer, acctsProcessor, indexCloner)
}

func getReindexerDataProcessor(cfg *config.Config, indicesConfigPath string) (DataProcessor, error) {
	sourceEsClient, err := elasticClient.NewElasticClient(cfg.Reindexer.SourceElasticSearchClient)
	if err != nil {
		return nil, err
	}

	destinationESClients, err := createESClients(cfg)
	if err != nil {
		return nil, err
	}

	rClient, err := restClient.NewRestClient(cfg.APIConfig.URL)
	if err != nil {
		return nil, err
	}

	pubKeyConverter, err := factory.NewPubkeyConverter(cfg.AddressPubkeyConverter)
	if err != nil {
		return nil, err
	}

	authenticationData := core.FetchAuthenticationData(cfg.APIConfig)
	acctGetter, err := NewAccountsGetter(rClient, cfg.GeneralConfig.DelegationLegacyContractAddress, pubKeyConverter, authenticationData, cfg.GeneralConfig.LKMEXStakingContractAddress)

	acctsProcessor, err := NewAccountsProcessor(rClient, acctGetter)
	if err != nil {
		return nil, err
	}

	reindexerProc, err := reindexer.New(sourceEsClient, destinationESClients, indicesConfigPath)
	if err != nil {
		return nil, err
	}

	return NewReindexerDataProcessor(acctsProcessor, reindexerProc)
}

func createESClients(cfg *config.Config) ([]crossIndex.ElasticClientHandler, error) {
	if len(cfg.Destination.DestinationElasticSearchClients) == 0 {
		return nil, errors.New("empty destination clients array")
	}

	clients := make([]crossIndex.ElasticClientHandler, 0)
	for _, esCfg := range cfg.Destination.DestinationElasticSearchClients {
		client, err := elasticClient.NewElasticClient(esCfg)
		if err != nil {
			return nil, err
		}

		clients = append(clients, client)
	}

	return clients, nil
}
