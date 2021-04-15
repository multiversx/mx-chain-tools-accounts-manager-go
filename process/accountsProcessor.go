package process

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-accounts-manager/convert"
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
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
func (ap *accountsProcessor) GetAllAccountsWithStake() (map[string]*data.AccountInfoWithStakeValues, []string, error) {
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

	calculateTotalStakeForAccounts(allAccounts)

	return allAccounts, allAddresses, nil
}

func calculateTotalStakeForAccounts(accounts map[string]*data.AccountInfoWithStakeValues) {
	for _, account := range accounts {
		totalStake, totalStakeNum := computeTotalBalance(
			account.DelegationLegacyWaiting,
			account.DelegationLegacyActive,
			account.ValidatorsActive,
			account.ValidatorTopUp,
			account.Delegation,
		)

		account.TotalStake = totalStake
		account.TotalStakeNum = totalStakeNum
	}
}

func (ap *accountsProcessor) mergeAccounts(
	legacyDelegators, validators, delegators map[string]*data.AccountInfoWithStakeValues,
) (map[string]*data.AccountInfoWithStakeValues, []string) {
	allAddresses := make([]string, 0)
	mergedAccounts := make(map[string]*data.AccountInfoWithStakeValues)

	for address, legacyDelegator := range legacyDelegators {
		mergedAccounts[address] = legacyDelegator

		allAddresses = append(allAddresses, address)
	}

	for address, stakedValidators := range validators {
		_, ok := mergedAccounts[address]
		if !ok {
			mergedAccounts[address] = stakedValidators

			allAddresses = append(allAddresses, address)
			continue
		}

		mergedAccounts[address].ValidatorsActive = stakedValidators.ValidatorsActive
		mergedAccounts[address].ValidatorsActiveNum = stakedValidators.ValidatorsActiveNum
		mergedAccounts[address].ValidatorTopUp = stakedValidators.ValidatorTopUp
		mergedAccounts[address].ValidatorTopUpNum = stakedValidators.ValidatorTopUpNum

		allAddresses = append(allAddresses, address)
	}

	for address, stakedDelegators := range delegators {
		_, ok := mergedAccounts[address]
		if !ok {
			mergedAccounts[address] = stakedDelegators

			allAddresses = append(allAddresses, address)
			continue
		}

		mergedAccounts[address].Delegation = stakedDelegators.Delegation
		mergedAccounts[address].DelegationNum = stakedDelegators.DelegationNum

		allAddresses = append(allAddresses, address)
	}

	return mergedAccounts, allAddresses
}

// PrepareAccountsForReindexing will prepare accounts for reindexing
func (ap *accountsProcessor) PrepareAccountsForReindexing(
	accountsES, accountsRest map[string]*data.AccountInfoWithStakeValues,
) map[string]*data.AccountInfoWithStakeValues {
	accounts := make(map[string]*data.AccountInfoWithStakeValues)

	for address, account := range accountsES {
		accounts[address] = account
	}

	for address, accountRest := range accountsRest {
		_, ok := accounts[address]
		if !ok {
			// this should never happen because accountsES and accountsRest should have same addresses
			accounts[address] = &data.AccountInfoWithStakeValues{}
		}

		accounts[address].StakeInfo = accountRest.StakeInfo

		totalBalanceWithStake, totalBalanceWithStakeNum := computeTotalBalance(
			accounts[address].Balance,
			accountRest.TotalStake,
		)

		accounts[address].TotalBalanceWithStake = totalBalanceWithStake
		accounts[address].TotalBalanceWithStakeNum = totalBalanceWithStakeNum
	}

	return accounts
}

// PrepareAccountsForReindexing will compute cloned accounts index based on current epoch
func (ap *accountsProcessor) ComputeClonedAccountsIndex() (string, error) {
	genericAPIResponse := &data.GenericAPIResponse{}
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

func computeTotalBalance(balances ...string) (string, float64) {
	totalBalance := big.NewInt(0)
	totalBalanceFloat := float64(0)

	if len(balances) == 0 {
		return "0", 0
	}

	for _, balance := range balances {
		balanceBig, ok := big.NewInt(0).SetString(balance, 10)
		if !ok {
			continue
		}

		totalBalance = totalBalance.Add(totalBalance, balanceBig)
		totalBalanceFloat += convert.ComputeBalanceAsFloat(balance)
	}

	return totalBalance.String(), totalBalanceFloat
}
