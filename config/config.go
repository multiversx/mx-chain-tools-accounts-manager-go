package config

import (
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

// Config will hold the whole config file's data
type Config struct {
	GeneralConfig          GeneralConfig
	AddressPubkeyConverter struct {
		Length int
		Type   string
	}
	Reindexer struct {
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
	EnergyContractAddress           string
	ValidatorsContract              string
}

// APIConfig holds the configuration for the API
type APIConfig struct {
	URL      string
	Username string
	Password string
}
