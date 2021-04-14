package data

import (
	"encoding/json"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/data/vm"
)

// GenericAPIResponse defines the structure of all responses on API endpoints
type GenericAPIResponse struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
	Code  string          `json:"code"`
}

// StakedInfo defines the structure of a response staked info response
type StakedInfo struct {
	Address string `json:"address"`
	Staked  string `json:"staked"`
	TopUp   string `json:"topUp"`
	Total   string `json:"total"`
}

type DelegatorStake struct {
	DelegatorAddress string `json:"delegatorAddress"`
	DelegatedTo      []struct {
		DelegationScAddress string `json:"delegatorAddress"`
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
	ID      string           `json:"_id"`
	Found   bool             `json:"found"`
	Account data.AccountInfo `json:"_source"`
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