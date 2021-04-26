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
	Reindexer              ReindexerConfig
	ApiCredentials         data.RestApiAuthenticationData
}

// GeneralSettingsConfig will hold the general settings for an accounts manager
type GeneralConfig struct {
	APIUrl                          string
	DelegationLegacyContractAddress string
}

// ClonerConfig holds the configuration necessary for a clone based indexer
type ClonerConfig struct {
	ElasticSearchClient data.EsClientConfig
}

// ReindexerConfig holds the configuration for a reindexer
type ReindexerConfig struct {
	SourceElasticSearchConfig     data.EsClientConfig
	DestinationElasticSeachConfig data.EsClientConfig
}
