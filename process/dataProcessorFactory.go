package process

import (
	"github.com/ElrondNetwork/elrond-accounts-manager/config"
	"github.com/ElrondNetwork/elrond-accounts-manager/elasticClient"
	"github.com/ElrondNetwork/elrond-accounts-manager/restClient"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/elastic/go-elasticsearch/v7"
)

// CreateDataProcessor will create a new instance of a data processor
func CreateDataProcessor(cfg *config.Config) (*dataProcessor, error) {
	elasticCfg := elasticsearch.Config{
		Addresses: []string{cfg.GeneralConfig.ElasticDatabaseAddress},
		Username:  cfg.GeneralConfig.Username,
		Password:  cfg.GeneralConfig.Password,
	}
	esClient, err := elasticClient.NewElasticClient(elasticCfg)
	if err != nil {
		return nil, err
	}

	rClient, err := restClient.NewRestClient(cfg.GeneralConfig.APIUrl)
	if err != nil {
		return nil, err
	}

	pubKeyConverter, err := factory.NewPubkeyConverter(cfg.AddressPubkeyConverter)
	if err != nil {
		return nil, err
	}

	acctGetter, err := newAccountsGetter(rClient, cfg.GeneralConfig.DelegationLegacyContractAddress, pubKeyConverter)

	acctsProcessor, err := NewAccountsProcessor(rClient, acctGetter)
	if err != nil {
		return nil, err
	}

	acctsIndexer, err := NewAccountsIndexer(esClient)
	if err != nil {
		return nil, err
	}

	indexCloner, err := NewCloner(esClient)
	if err != nil {
		return nil, err
	}

	return NewDataProcessor(acctsIndexer, acctsProcessor, indexCloner)
}
