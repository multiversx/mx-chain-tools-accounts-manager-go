package process

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	dataTypes "github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/tidwall/gjson"
)

const (
	pathNodeStatusMeta = "/network/status/4294967295"
)

type accountsProcessor struct {
	accountsGetterHandler
	restClient RestClientHandler
}

// NewAccountsProcessor will create a new instance of accountsProcessor
func NewAccountsProcessor(restClient RestClientHandler, acctsGetter accountsGetterHandler) (*accountsProcessor, error) {
	return &accountsProcessor{
		restClient:            restClient,
		accountsGetterHandler: acctsGetter,
	}, nil
}

// GetAllAccountsWithStake will return all accounts with stake
func (ap *accountsProcessor) GetAllAccountsWithStake() (map[string]*data.AccountInfo, []string, error) {
	legacyDelegators, err := ap.getLegacyDelegatorsAccounts()
	if err != nil {
		return nil, nil, err
	}

	validators, err := ap.getValidatorsAccounts()
	if err != nil {
		return nil, nil, err
	}

	delegators, err := ap.getDelegatorsAccounts()
	if err != nil {
		return nil, nil, err
	}

	allAccounts, allAddresses := ap.mergeAccounts(legacyDelegators, validators, delegators)

	return allAccounts, allAddresses, nil
}

func (ap *accountsProcessor) mergeAccounts(legacyDelegators, validators, delegators map[string]string) (map[string]*data.AccountInfo, []string) {
	allAddresses := make([]string, 0)
	mergedAccounts := make(map[string]*data.AccountInfo)

	for address, stakedLegacyDelegation := range legacyDelegators {
		mergedAccounts[address] = &data.AccountInfo{
			ValueStakedLegacyDelegation: stakedLegacyDelegation,
		}

		allAddresses = append(allAddresses, address)
	}

	for address, stakedValidators := range validators {
		_, ok := mergedAccounts[address]
		if !ok {
			mergedAccounts[address] = &data.AccountInfo{}
		}

		mergedAccounts[address].ValueStakedValidators = stakedValidators
		allAddresses = append(allAddresses, address)
	}

	for address, stakedDelegators := range delegators {
		_, ok := mergedAccounts[address]
		if !ok {
			mergedAccounts[address] = &data.AccountInfo{}
		}

		mergedAccounts[address].ValueStakedDelegation = stakedDelegators
		allAddresses = append(allAddresses, address)
	}

	return mergedAccounts, allAddresses
}

// PrepareAccountsForReindexing will prepare accounts for reindexing
func (ap *accountsProcessor) PrepareAccountsForReindexing(accountsES, accountsRest map[string]*data.AccountInfo) map[string]*data.AccountInfo {
	accounts := make(map[string]*data.AccountInfo)

	for address, account := range accountsES {
		accounts[address] = account
	}

	for address, account := range accountsRest {
		_, ok := accounts[address]
		if !ok {
			// this should never happen because accountsES and accountsRest should have same addresses
			accounts[address] = &data.AccountInfo{}
		}

		accounts[address].TotalBalanceWithStake = computeTotalBalance(
			accounts[address].Balance,
			account.ValueStakedDelegation,
			account.ValueStakedLegacyDelegation,
			account.ValueStakedValidators,
		)
	}

	return accounts
}

// PrepareAccountsForReindexing will compute cloned accounts index based on current epoch
func (ap *accountsProcessor) ComputeClonedAccountsIndex() (string, error) {
	genericAPIResponse := &dataTypes.GenericAPIResponse{}
	err := ap.restClient.CallGetRestEndPoint(pathNodeStatusMeta, genericAPIResponse)
	if err != nil {
		return "", err
	}
	if genericAPIResponse.Error != "" {
		return "", fmt.Errorf("%s", genericAPIResponse.Error)
	}

	epoch := gjson.Get(string(genericAPIResponse.Data), "status.erd_epoch_number")

	return fmt.Sprintf("%s_%s", accountsIndex, epoch.String()), nil
}

func computeTotalBalance(balances ...string) string {
	totalBalance := big.NewInt(0)

	if len(balances) == 0 {
		return "0"
	}

	for _, balance := range balances {
		balanceBig, ok := big.NewInt(0).SetString(balance, 10)
		if !ok {
			continue
		}

		totalBalance = totalBalance.Add(totalBalance, balanceBig)
	}

	return totalBalance.String()
}
