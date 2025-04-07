package process

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	nodeCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/config"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/multiversx/mx-chain-vm-common-go"
	"github.com/tidwall/gjson"
)

const (
	pathValidatorsStake = "/network/direct-staked-info"
	pathDelegatorStake  = "/network/delegated-info"
	pathVMValues        = "/vm-values/query"
	lkMexSnapShot       = "getSnapshot"
	pathAccountKeys     = "/address/%s/keys"
)

type accountsGetter struct {
	unDelegatedInfoProc *unDelegatedInfoProcessor
	restClient          RestClientHandler
	pubKeyConverter     nodeCore.PubkeyConverter
	authenticationData  data.RestApiAuthenticationData
	mutex               sync.Mutex

	delegationContractAddress string
	lkMexContractAddress      string
	energyContractAddress     string
	validatorsContract        string
}

// NewAccountsGetter will create a new instance of accountsGetter
func NewAccountsGetter(
	restClient RestClientHandler,
	pubKeyConverter nodeCore.PubkeyConverter,
	authenticationData data.RestApiAuthenticationData,
	generalConfig config.GeneralConfig,
	esClient ElasticClientHandler,
) (*accountsGetter, error) {
	return &accountsGetter{
		mutex:                     sync.Mutex{},
		restClient:                restClient,
		pubKeyConverter:           pubKeyConverter,
		authenticationData:        authenticationData,
		lkMexContractAddress:      generalConfig.LKMEXStakingContractAddress,
		energyContractAddress:     generalConfig.EnergyContractAddress,
		delegationContractAddress: generalConfig.DelegationLegacyContractAddress,
		validatorsContract:        generalConfig.ValidatorsContract,
		unDelegatedInfoProc:       newUnDelegateInfoProcessor(esClient),
	}, nil
}

// GetLegacyDelegatorsAccounts will fetch all accounts with stake from API
func (ag *accountsGetter) GetLegacyDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	defer logExecutionTime(time.Now(), "Fetched accounts from legacy delegation contract")

	responseKeys := &data.GenericAPIResponse{}
	path := fmt.Sprintf(pathAddressKeys, ag.delegationContractAddress)
	err := ag.restClient.CallGetRestEndPoint(path, responseKeys, core.GetEmptyApiCredentials())
	if err != nil {
		return nil, err
	}
	if responseKeys.Error != "" {
		return nil, fmt.Errorf("%s", responseKeys.Error)
	}

	pairs := gjson.Get(string(responseKeys.Data), "pairs")

	pairsMap := make(map[string]string)
	err = json.Unmarshal([]byte(pairs.String()), &pairsMap)
	if err != nil {
		log.Warn("cannot unmarshal accounts info", "error", err.Error())
		return nil, err
	}

	return ag.extractDelegationLegacyData(pairsMap)
}

// GetValidatorsAccounts will fetch all validators accounts
func (ag *accountsGetter) GetValidatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	defer logExecutionTime(time.Now(), "Fetched accounts from validators contract")

	genericApiResponse := &data.GenericAPIResponse{}
	err := ag.restClient.CallGetRestEndPoint(pathValidatorsStake, genericApiResponse, ag.authenticationData)
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

	accountsStake := make(map[string]*data.AccountInfoWithStakeValues)
	for _, acct := range accountsInfo {
		accountsStake[acct.Address] = &data.AccountInfoWithStakeValues{
			StakeInfo: data.StakeInfo{
				ValidatorsActive:    acct.Staked,
				ValidatorsActiveNum: core.ComputeBalanceAsFloat(acct.Staked),
				ValidatorTopUp:      acct.TopUp,
				ValidatorTopUpNum:   core.ComputeBalanceAsFloat(acct.TopUp),
			},
		}
	}

	log.Info("validators accounts", "num", len(accountsStake))

	err = ag.putUndelegatedValuesFromValidatorsContract(accountsStake)
	if err != nil {
		return nil, err
	}

	return accountsStake, nil
}

// GetDelegatorsAccounts will fetch all delegators accounts
func (ag *accountsGetter) GetDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	defer logExecutionTime(time.Now(), "Fetched accounts from delegation manager contracts")

	genericApiResponse := &data.GenericAPIResponse{}
	err := ag.restClient.CallGetRestEndPoint(pathDelegatorStake, genericApiResponse, ag.authenticationData)
	if err != nil {
		log.Warn("CallGetRestEndPoint", "error", err.Error())
		return nil, err
	}
	if genericApiResponse.Error != "" {
		return nil, fmt.Errorf("cannot get delegators accounts %s", genericApiResponse.Error)
	}

	list := gjson.Get(string(genericApiResponse.Data), "list")
	accountsInfo := make([]data.DelegatorStake, 0)
	err = json.Unmarshal([]byte(list.String()), &accountsInfo)
	if err != nil {
		log.Warn("cannot unmarshal accounts info", "error", err.Error())
		return nil, err
	}

	accountsStake := make(map[string]*data.AccountInfoWithStakeValues)
	for _, acct := range accountsInfo {
		accountsStake[acct.DelegatorAddress] = &data.AccountInfoWithStakeValues{
			StakeInfo: data.StakeInfo{
				Delegation:    acct.Total,
				DelegationNum: core.ComputeBalanceAsFloat(acct.Total),
			},
		}
	}

	log.Info("delegators accounts", "num", len(accountsStake))

	err = ag.unDelegatedInfoProc.putUnDelegateInfoFromStakingProviders(accountsStake)
	if err != nil {
		return nil, err
	}

	return accountsStake, nil
}

// GetLKMEXStakeAccounts will fetch all accounts that have stake lkmex tokens
func (ag *accountsGetter) GetLKMEXStakeAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	accountsMap := make(map[string]*data.AccountInfoWithStakeValues)
	if ag.lkMexContractAddress == "" {
		return accountsMap, nil
	}

	defer logExecutionTime(time.Now(), "Fetched accounts from lkmex staking contract")

	vmRequest := &data.VmValueRequest{
		Address:    ag.lkMexContractAddress,
		FuncName:   lkMexSnapShot,
		CallerAddr: ag.lkMexContractAddress,
	}

	responseVmValue := &data.ResponseVmValue{}
	err := ag.restClient.CallPostRestEndPoint(pathVMValues, vmRequest, responseVmValue, core.GetEmptyApiCredentials())
	if err != nil {
		return nil, err
	}
	if responseVmValue.Error != "" {
		return nil, fmt.Errorf("%s", responseVmValue.Error)
	}
	if responseVmValue.Data.Data != nil {
		if responseVmValue.Data.Data.ReturnCode != vmcommon.Ok.String() {
			return nil, fmt.Errorf("%s: %s", responseVmValue.Data.Data.ReturnCode, responseVmValue.Data.Data.ReturnMessage)
		}
	}

	stepForLoop := 2
	returnedData := responseVmValue.Data.Data.ReturnData
	accountsStake := make(map[string]string, 0)
	for idx := 0; idx < len(returnedData); idx += stepForLoop {
		address := ag.pubKeyConverter.Encode(returnedData[idx])
		stakedBalance := big.NewInt(0).SetBytes(returnedData[idx+1])

		accountsStake[address] = stakedBalance.String()
	}

	for key, value := range accountsStake {
		accountsMap[key] = &data.AccountInfoWithStakeValues{
			StakeInfo: data.StakeInfo{
				LKMEXStake:    value,
				LKMEXStakeNum: core.ComputeBalanceAsFloat(value),
			},
		}
	}

	log.Info("staked lkmex accounts", "num", len(accountsStake))

	return accountsMap, nil
}

func logExecutionTime(start time.Time, message string) {
	log.Info(message, "duration in seconds", time.Since(start).Seconds())
}
