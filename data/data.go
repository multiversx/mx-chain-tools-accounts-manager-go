package data

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// GenericAPIResponse defines the structure of all responses on API endpoints
type GenericAPIResponse struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
	Code  string          `json:"code"`
}

// BlockInfo defines the structure of block info
type BlockInfo struct {
	Hash     string `json:"hash"`
	Nonce    uint64 `json:"nonce"`
	RootHash string `json:"rootHash"`
}

// StakedInfo defines the structure of a response staked info response
type StakedInfo struct {
	Address string `json:"address"`
	Staked  string `json:"baseStaked"`
	TopUp   string `json:"topUp"`
	Total   string `json:"total"`
}

type DelegatorStake struct {
	DelegatorAddress string `json:"delegatorAddress"`
	DelegatedTo      []struct {
		DelegationScAddress string `json:"delegationScAddress"`
		Value               string `json:"value"`
	} `json:"delegatedTo"`
	Total string `json:"total"`
}

// VmValuesResponseData follows the format of the data field in an API response for a VM values query
type VmValuesResponseData struct {
	Data *vm.VMOutputApi `json:"data"`
}

// ResponseVmValue defines a wrapper over string containing returned data in hex format
type ResponseVmValue struct {
	Data  VmValuesResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

// VmValueRequest defines the request struct for values available in a VM
type VmValueRequest struct {
	Address    string   `json:"scAddress"`
	FuncName   string   `json:"funcName"`
	CallerAddr string   `json:"caller"`
	CallValue  string   `json:"value"`
	Args       []string `json:"args"`
}

// SCQuery represents a prepared query for executing a function of the smart contract
type SCQuery struct {
	ScAddress  string
	FuncName   string
	CallerAddr string
	CallValue  string
	Arguments  [][]byte
}

// AccountsResponseES defines the structure of a response
type AccountsResponseES struct {
	ID      string                     `json:"_id"`
	Found   bool                       `json:"found"`
	Account AccountInfoWithStakeValues `json:"_source"`
}

// BulkRequestResponse defines the structure of a bulk request response
type BulkRequestResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			Status int `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

// AccountInfoWithStakeValues extends the structure data.AccountInfo with stake values
type AccountInfoWithStakeValues struct {
	data.AccountInfo
	StakeInfo
}

type AccountsData struct {
	AccountsWithStake map[string]*AccountInfoWithStakeValues
	Addresses         []string
	EnergyBlockInfo   *BlockInfo
	Epoch             uint32
}

// StakeInfo is the structure that contains all information about stake for an account
type StakeInfo struct {
	DelegationLegacyWaiting    string  `json:"delegationLegacyWaiting,omitempty"`
	DelegationLegacyWaitingNum float64 `json:"delegationLegacyWaitingNum,omitempty"`
	DelegationLegacyActive     string  `json:"delegationLegacyActive,omitempty"`
	DelegationLegacyActiveNum  float64 `json:"delegationLegacyActiveNum,omitempty"`
	ValidatorsActive           string  `json:"validatorsActive,omitempty"`
	ValidatorsActiveNum        float64 `json:"validatorsActiveNum,omitempty"`
	ValidatorTopUp             string  `json:"validatorsTopUp,omitempty"`
	ValidatorTopUpNum          float64 `json:"validatorsTopUpNum,omitempty"`
	Delegation                 string  `json:"delegation,omitempty"`
	DelegationNum              float64 `json:"delegationNum,omitempty"`
	TotalStake                 string  `json:"totalStake,omitempty"`
	TotalStakeNum              float64 `json:"totalStakeNum,omitempty"`

	LKMEXStake    string         `json:"lkMexStake,omitempty"`
	LKMEXStakeNum float64        `json:"lkMexStakeNum,omitempty"`
	Energy        string         `json:"energy,omitempty"`
	EnergyNum     float64        `json:"energyNum,omitempty"`
	EnergyDetails *EnergyDetails `json:"energyDetails,omitempty"`

	UnDelegateLegacy        string  `json:"unDelegateLegacy,omitempty"`
	UnDelegateLegacyNum     float64 `json:"unDelegateLegacyNum,omitempty"`
	UnDelegateValidator     string  `json:"unDelegateValidator,omitempty"`
	UnDelegateValidatorNum  float64 `json:"unDelegateValidatorNum,omitempty"`
	UnDelegateDelegation    string  `json:"unDelegateDelegation,omitempty"`
	UnDelegateDelegationNum float64 `json:"unDelegateDelegationNum,omitempty"`
	TotalUnDelegate         string  `json:"totalUnDelegate,omitempty"`
	TotalUnDelegateNum      float64 `json:"totalUnDelegateNum,omitempty"`
}

// EnergyDetails is the structure that contains details about the user's energy
type EnergyDetails struct {
	LastUpdateEpoch   uint32 `json:"lastUpdateEpoch"`
	Amount            string `json:"amount"`
	TotalLockedTokens string `json:"totalLockedTokens"`
}

// KeyValueObj is the dto for values index
type KeyValueObj struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// EsClientConfig is a wrapper over the internally used field from elasticsearch.Config struct
type EsClientConfig struct {
	Address  string
	Username string
	Password string
}

// RestApiAuthenticationData holds the data to be used when authorizing API requests
type RestApiAuthenticationData struct {
	Username string
	Password string
}
