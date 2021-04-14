package process

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/tidwall/gjson"
)

const (
	pathValidatorsStake = "/network/direct-staked-info"
	pathDelegatorStake  = "/network/delegated-info"
	pathVMValues        = "/vm-values/query"
	getFullWaitingList  = "getFullWaitingList"
	getFullActiveList   = "getFullActiveList"
)

type accountsGetter struct {
	restClient                RestClientHandler
	delegationContractAddress string
	pubKeyConverter           core.PubkeyConverter
}

func newAccountsGetter(
	restClient RestClientHandler,
	delegationContractAddress string,
	pubKeyConverter core.PubkeyConverter,
) (*accountsGetter, error) {
	return &accountsGetter{
		restClient:                restClient,
		delegationContractAddress: delegationContractAddress,
		pubKeyConverter:           pubKeyConverter,
	}, nil
}

func (ag *accountsGetter) getLegacyDelegatorsAccounts() (map[string]string, error) {
	defer logExecutionTime(time.Now(), "Fetched accounts from legacy delegation contract")

	activeListAccounts, err := ag.getFullActiveListAccounts()
	if err != nil {
		return nil, err
	}

	fullWaitingListAccounts, err := ag.getFullWaitingListAccounts()
	if err != nil {
		return nil, err
	}

	mergedMaps := make(map[string]string)
	for key, value := range activeListAccounts {
		mergedMaps[key] = value
	}

	for key, value := range fullWaitingListAccounts {
		_, ok := mergedMaps[key]
		if !ok {
			mergedMaps[key] = value
			continue
		}

		mergedMaps[key] = computeTotalBalance(mergedMaps[key], value)
	}

	return mergedMaps, nil
}

func (ag *accountsGetter) getFullActiveListAccounts() (map[string]string, error) {
	return ag.getAccountsVMQuery(getFullActiveList, 2)
}

func (ag *accountsGetter) getFullWaitingListAccounts() (map[string]string, error) {
	return ag.getAccountsVMQuery(getFullWaitingList, 3)
}

func (ag *accountsGetter) getAccountsVMQuery(funcName string, stepForLoop int) (map[string]string, error) {
	vmRequest := &data.VmValueRequest{
		Address:    ag.delegationContractAddress,
		FuncName:   funcName,
		CallerAddr: ag.delegationContractAddress,
	}

	responseVmValue := &data.ResponseVmValue{}
	err := ag.restClient.CallPostRestEndPoint(pathVMValues, vmRequest, responseVmValue)
	if err != nil {
		return nil, err
	}
	if responseVmValue.Error != "" {
		return nil, fmt.Errorf("%s", responseVmValue.Error)
	}

	returnedData := responseVmValue.Data.Data.ReturnData
	accountsStake := make(map[string]string, 0)
	for idx := 0; idx < len(returnedData); idx += stepForLoop {
		address := ag.pubKeyConverter.Encode(returnedData[idx])
		stakedBalance := big.NewInt(0).SetBytes(returnedData[idx+1])

		accountsStake[address] = stakedBalance.String()
	}

	return accountsStake, nil
}

func (ag *accountsGetter) getValidatorsAccounts() (map[string]string, error) {
	defer logExecutionTime(time.Now(), "Fetched accounts from validators contract")

	genericApiResponse := &data.GenericAPIResponse{}
	err := ag.restClient.CallGetRestEndPoint(pathValidatorsStake, genericApiResponse)
	if err != nil {
		return nil, err
	}
	if genericApiResponse.Error != "" {
		return nil, fmt.Errorf("%s", genericApiResponse.Error)
	}

	list := gjson.Get(string(genericApiResponse.Data), "list")
	accountsInfo := make([]data.StakedInfo, 0)
	err = json.Unmarshal([]byte(list.String()), &accountsInfo)
	if err != nil {
		return nil, err
	}

	accountsStake := make(map[string]string, len(accountsInfo))
	for _, acct := range accountsInfo {
		accountsStake[acct.Address] = acct.Total
	}

	return accountsStake, nil
}

func (ag *accountsGetter) getDelegatorsAccounts() (map[string]string, error) {
	defer logExecutionTime(time.Now(), "Fetched accounts from delegation manager contracts")

	genericApiResponse := &data.GenericAPIResponse{}
	err := ag.restClient.CallGetRestEndPoint(pathDelegatorStake, genericApiResponse)
	if err != nil {
		return nil, err
	}
	if genericApiResponse.Error != "" {
		return nil, fmt.Errorf("%s", genericApiResponse.Error)
	}

	list := gjson.Get(string(genericApiResponse.Data), "list")
	accountsInfo := make([]data.DelegatorStake, 0)
	err = json.Unmarshal([]byte(list.String()), &accountsInfo)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	accountsStake := make(map[string]string, len(accountsInfo))
	for _, acct := range accountsInfo {
		accountsStake[acct.DelegatorAddress] = acct.Total
	}

	return accountsStake, nil
}