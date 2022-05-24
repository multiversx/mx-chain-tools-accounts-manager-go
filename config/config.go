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
	Reindexer              struct {
		SourceElasticSearchClient data.EsClientConfig
	}
	Destination struct {
		DestinationElasticSearchClients []data.EsClientConfig `toml:"DestinationElasticSearchClients"`
	}
	APIConfig APIConfig
}

// GeneralConfig will hold the general settings for an accounts manager
type GeneralConfig struct {
	DelegationLegacyContractAddress string
	LKMEXStakingContractAddress     string
}

// ClonerConfig holds the configuration necessary for a clone based indexer
type ClonerConfig struct {
	ElasticSearchClient data.EsClientConfig
}

// APIConfig holds the configuration for the API
type APIConfig struct {
	URL      string
	Username string
	Password string
}
