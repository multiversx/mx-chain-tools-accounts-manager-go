package config

import (
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-go/config"
)

// Config will hold the whole config file's data
type Config struct {
	GeneralConfig          GeneralConfig
	AddressPubkeyConverter config.PubkeyConfig
	Cloner                 ClonerConfig
	Reindexer              Reindexer
	APIConfig              APIConfig
}

// GeneralSettingsConfig will hold the general settings for an accounts manager
type GeneralConfig struct {
	DelegationLegacyContractAddress string
}

// ClonerConfig holds the configuration necessary for a clone based indexer
type ClonerConfig struct {
	ElasticSearchClient data.EsClientConfig
}

// Reindexer holds the configuration for a reindexer
type Reindexer struct {
	SourceElasticSearchClient      data.EsClientConfig
	DestinationElasticSearchClient data.EsClientConfig
}

// APIConfig holds the configuration for the API
type APIConfig struct {
	URL      string
	Username string
	Password string
}
