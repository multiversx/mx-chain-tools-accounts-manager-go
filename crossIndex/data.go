package crossIndex

import (
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/mappings"
)

const AccountsPolicyName = "accounts-manager-retention-policy"

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

// AccountsTemplate will hold the configuration for the accounts index
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
	"settings": mappings.Object{
		"number_of_shards":     1,
		"number_of_replicas":   1,
		"index.lifecycle.name": AccountsPolicyName,
	},
}

var AccountsClonedPolicy = mappings.Object{
	"policy": mappings.Object{
		"phases": mappings.Object{
			"hot": mappings.Object{
				"min_age": "0ms",
				"actions": mappings.Object{},
			},
			"delete": mappings.Object{
				"min_age": "90d",
				"actions": mappings.Object{
					"delete": mappings.Object{
						"delete_searchable_snapshot": true,
					},
				},
			},
		},
	},
}
