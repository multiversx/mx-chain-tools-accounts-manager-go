package config

import "github.com/ElrondNetwork/elrond-go/config"

// Config will hold the whole config file's data
type Config struct {
	GeneralConfig          GeneralConfig
	AddressPubkeyConverter config.PubkeyConfig
}

// GeneralSettingsConfig will hold the general settings for an accounts manager
type GeneralConfig struct {
	APIUrl                          string
	ElasticDatabaseAddress          string
	Username                        string
	Password                        string
	DelegationLegacyContractAddress string
}
