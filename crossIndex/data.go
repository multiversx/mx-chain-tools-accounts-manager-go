package crossIndex

import (
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

// AccountsPolicyName is the name of the policy for the accounts index
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
