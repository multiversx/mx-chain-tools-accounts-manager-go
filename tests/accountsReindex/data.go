package accountsReindex

import (
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/mappings"
)

type AllAccountsResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []struct {
			ID      string                          `json:"_id"`
			Account data.AccountInfoWithStakeValues `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// AccountsTemplate will hold the configuration for the accounts index
var AccountsTemplate = mappings.Object{
	"index_patterns": []interface{}{
		"accounts-*",
	},
	"settings": mappings.Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},

	"mappings": mappings.Object{
		"properties": mappings.Object{
			"balanceNum": mappings.Object{
				"type": "double",
			},
			"totalBalanceWithStakeNum": mappings.Object{
				"type": "double",
			},
		},
	},
}
