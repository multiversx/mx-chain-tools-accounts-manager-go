package process

import (
	"github.com/ElrondNetwork/elrond-accounts-manager/config"
	"github.com/ElrondNetwork/elrond-accounts-manager/core"
	"github.com/ElrondNetwork/elrond-accounts-manager/crossIndex/cloner"
	"github.com/ElrondNetwork/elrond-accounts-manager/crossIndex/reindexer"
	"github.com/ElrondNetwork/elrond-accounts-manager/elasticClient"
	"github.com/ElrondNetwork/elrond-accounts-manager/process/accountsIndexer"
	"github.com/ElrondNetwork/elrond-accounts-manager/restClient"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
)

// CreateDataProcessor will create a new instance of a data processor
func CreateDataProcessor(cfg *config.Config, indexType string) (DataProcessor, error) {
	if indexType == "clone" {
		return getClonerDataProcessor(cfg)
	}

	return getReindexerDataProcessor(cfg)
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
	acctGetter, err := NewAccountsGetter(rClient, cfg.GeneralConfig.DelegationLegacyContractAddress, pubKeyConverter, authenticationData)

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

func getReindexerDataProcessor(cfg *config.Config) (DataProcessor, error) {
	sourceEsClient, err := elasticClient.NewElasticClient(cfg.Reindexer.SourceElasticSearchClient)
	if err != nil {
		return nil, err
	}

	destinationEsClient, err := elasticClient.NewElasticClient(cfg.Reindexer.DestinationElasticSearchClient)
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
	acctGetter, err := NewAccountsGetter(rClient, cfg.GeneralConfig.DelegationLegacyContractAddress, pubKeyConverter, authenticationData)

	acctsProcessor, err := NewAccountsProcessor(rClient, acctGetter)
	if err != nil {
		return nil, err
	}

	reindexerProc, err := reindexer.New(sourceEsClient, destinationEsClient)
	if err != nil {
		return nil, err
	}

	return NewReindexerDataProcessor(acctsProcessor, reindexerProc)
}
