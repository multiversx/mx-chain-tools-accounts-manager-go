package crossIndex

import (
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/mappings"
)

// AllAccountsResponse is a structure that matches the response format for an all accounts request
type AllAccountsResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []struct {
			ID      string                          `json:"_id"`
			Account data.AccountInfoWithStakeValues `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// Accounts will hold the configuration for the accounts index
var AccountsTemplate = mappings.Object{
	"mappings": mappings.Object{
		"properties": mappings.Object{
			"balanceNum": mappings.Object{
				"type": "double",
			},
			"delegationLegacyWaitingNum": mappings.Object{
				"type": "double",
			},
			"delegationLegacyActiveNum": mappings.Object{
				"type": "double",
			},
			"validatorsActiveNum": mappings.Object{
				"type": "double",
			},
			"validatorsTopUpNum": mappings.Object{
				"type": "double",
			},
			"delegationNum": mappings.Object{
				"type": "double",
			},
			"totalStakeNum": mappings.Object{
				"type": "double",
			},
			"totalBalanceWithStakeNum": mappings.Object{
				"type": "double",
			},
		},
	},
}
