package process

import (
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/tidwall/gjson"
)

const (
	pathNodeStatusMeta = "/network/status/4294967295"
)

type accountsProcessor struct {
	AccountsGetterHandler
	restClient RestClientHandler
}

// NewAccountsProcessor will create a new instance of accountsProcessor
func NewAccountsProcessor(restClient RestClientHandler, acctsGetter AccountsGetterHandler) (*accountsProcessor, error) {
	return &accountsProcessor{
		restClient:            restClient,
		AccountsGetterHandler: acctsGetter,
	}, nil
}

// GetAllAccountsWithStake will return all accounts with stake
func (ap *accountsProcessor) GetAllAccountsWithStake(currentEpoch uint32) (*data.AccountsData, error) {
	legacyDelegators, err := ap.GetLegacyDelegatorsAccounts()
	if err != nil {
		return nil, err
	}

	validators, err := ap.GetValidatorsAccounts()
	if err != nil {
		return nil, err
	}

	delegators, err := ap.GetDelegatorsAccounts()
	if err != nil {
		return nil, err
	}

	lkMexAccountsWithStake, err := ap.GetLKMEXStakeAccounts()
	if err != nil {
		return nil, err
	}

	accountsWithEnergy, blockInfoEnergy, err := ap.GetAccountsWithEnergy(currentEpoch)
	if err != nil {
		return nil, err
	}

	allAccounts, allAddresses := ap.mergeAccounts(legacyDelegators, validators, delegators, lkMexAccountsWithStake, accountsWithEnergy)

	calculateTotalStakeForAccountsAndTotalUnDelegated(allAccounts)

	return &data.AccountsData{
		AccountsWithStake: allAccounts,
		Addresses:         allAddresses,
		EnergyBlockInfo:   blockInfoEnergy,
		Epoch:             currentEpoch,
	}, nil
}

func calculateTotalStakeForAccountsAndTotalUnDelegated(accounts map[string]*data.AccountInfoWithStakeValues) {
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

		totalUnDelegated, totalUnDelegatedNum := computeTotalBalance(
			account.UnDelegateLegacy,
			account.UnDelegateValidator,
			account.UnDelegateDelegation,
		)

		account.TotalUnDelegate = totalUnDelegated
		account.TotalUnDelegateNum = totalUnDelegatedNum
	}
}

func (ap *accountsProcessor) mergeAccounts(
	legacyDelegators, validators, delegators, lkMexAccountsWithStake, accountsWithEnergy map[string]*data.AccountInfoWithStakeValues,
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

		mergedAccounts[address].UnDelegateValidator = stakedValidators.UnDelegateValidator
		mergedAccounts[address].UnDelegateValidatorNum = stakedValidators.UnDelegateValidatorNum
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

		mergedAccounts[address].UnDelegateDelegation = stakedDelegators.UnDelegateDelegation
		mergedAccounts[address].UnDelegateDelegationNum = stakedDelegators.UnDelegateDelegationNum
	}

	for address, lkMexAccount := range lkMexAccountsWithStake {
		_, ok := mergedAccounts[address]
		if !ok {
			mergedAccounts[address] = lkMexAccount

			allAddresses = append(allAddresses, address)
			continue
		}

		mergedAccounts[address].LKMEXStake = lkMexAccount.LKMEXStake
		mergedAccounts[address].LKMEXStakeNum = lkMexAccount.LKMEXStakeNum
	}

	for address, energyAccount := range accountsWithEnergy {
		_, ok := mergedAccounts[address]
		if !ok {
			mergedAccounts[address] = energyAccount

			allAddresses = append(allAddresses, address)
			continue
		}

		mergedAccounts[address].Energy = energyAccount.Energy
		mergedAccounts[address].EnergyNum = energyAccount.EnergyNum
		mergedAccounts[address].EnergyDetails = energyAccount.EnergyDetails
	}

	return mergedAccounts, allAddresses
}

// ComputeClonedAccountsIndex will compute cloned accounts index based on current epoch
func (ap *accountsProcessor) ComputeClonedAccountsIndex(epoch uint32) (string, error) {
	log.Info("Compute name of the new index...")

	return fmt.Sprintf("%s_%d", accountsIndex, epoch), nil
}

// GetCurrentEpoch will fetch the current epoch from the network
func (ap *accountsProcessor) GetCurrentEpoch() (uint32, error) {
	genericAPIResponse := &data.GenericAPIResponse{}
	err := ap.restClient.CallGetRestEndPoint(pathNodeStatusMeta, genericAPIResponse, core.GetEmptyApiCredentials())
	if err != nil {
		return 0, err
	}
	if genericAPIResponse.Error != "" {
		return 0, fmt.Errorf("cannot compute accounts index %s", genericAPIResponse.Error)
	}

	epoch := gjson.Get(string(genericAPIResponse.Data), "status.erd_epoch_number")
	return uint32(epoch.Num), nil
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
		totalBalanceFloat += core.ComputeBalanceAsFloat(balance)
	}

	return totalBalance.String(), totalBalanceFloat
}

// IsInterfaceNil returns true if the value under the interface is nil
func (ap *accountsProcessor) IsInterfaceNil() bool {
	return ap == nil
}
