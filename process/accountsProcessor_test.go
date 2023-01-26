package process

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/mocks"
	"github.com/stretchr/testify/require"
)

func TestAccountsProcessor_GetAllAccountsWithStake(t *testing.T) {
	t.Parallel()

	accountsDelegation := generateAccounts(delegation, 20)
	accountsDelegationLegacy := generateAccounts(delegationLegacy, 30)
	accountsValidators := generateAccounts(validator, 30)

	keys := make([]string, 45)
	for idx := 0; idx < 45; idx++ {
		keys[idx] = generateRandomBigIntString()
	}

	mapDelegation := makeMapFromArrays(keys[:20], accountsDelegation)
	mapLegacyDelegation := makeMapFromArrays(keys[5:35], accountsDelegationLegacy)
	mapValidators := makeMapFromArrays(keys[15:45], accountsValidators)

	ap, err := NewAccountsProcessor(&mocks.RestClientStub{}, &mocks.AccountsGetterStub{
		GetDelegatorsAccountsCalled: func() (map[string]*data.AccountInfoWithStakeValues, error) {
			return mapDelegation, nil
		},
		GetLegacyDelegatorsAccountsCalled: func() (map[string]*data.AccountInfoWithStakeValues, error) {
			return mapLegacyDelegation, nil
		},
		GetValidatorsAccountsCalled: func() (map[string]*data.AccountInfoWithStakeValues, error) {
			return mapValidators, nil
		},
	})
	require.Nil(t, err)

	accountsData, err := ap.GetAllAccountsWithStake(0)
	require.Nil(t, err)
	require.Equal(t, len(accountsData.AccountsWithStake), len(accountsData.Addresses))

	for addr, processedAccount := range accountsData.AccountsWithStake {
		acctDelegation, ok := mapDelegation[addr]
		if !ok {
			acctDelegation = &data.AccountInfoWithStakeValues{}
		}

		acctLegacyDelegation, ok := mapLegacyDelegation[addr]
		if !ok {
			acctLegacyDelegation = &data.AccountInfoWithStakeValues{}
		}

		acctValidator, ok := mapValidators[addr]
		if !ok {
			acctValidator = &data.AccountInfoWithStakeValues{}
		}

		require.Equal(t, acctDelegation.Delegation, processedAccount.Delegation)
		require.Equal(t, acctDelegation.DelegationNum, processedAccount.DelegationNum)

		require.Equal(t, acctLegacyDelegation.DelegationLegacyActive, processedAccount.DelegationLegacyActive)
		require.Equal(t, acctLegacyDelegation.DelegationLegacyActiveNum, processedAccount.DelegationLegacyActiveNum)
		require.Equal(t, acctLegacyDelegation.DelegationLegacyWaiting, processedAccount.DelegationLegacyWaiting)
		require.Equal(t, acctLegacyDelegation.DelegationLegacyWaitingNum, processedAccount.DelegationLegacyWaitingNum)

		require.Equal(t, acctValidator.ValidatorsActive, processedAccount.ValidatorsActive)
		require.Equal(t, acctValidator.ValidatorsActiveNum, processedAccount.ValidatorsActiveNum)
		require.Equal(t, acctValidator.ValidatorTopUp, processedAccount.ValidatorTopUp)
		require.Equal(t, acctValidator.ValidatorTopUpNum, processedAccount.ValidatorTopUpNum)

		expectedTotalStake, _ := computeTotalBalance(
			acctDelegation.Delegation,
			acctLegacyDelegation.DelegationLegacyActive,
			acctLegacyDelegation.DelegationLegacyWaiting,
			acctValidator.ValidatorsActive,
			acctValidator.ValidatorTopUp,
		)
		require.Equal(t, expectedTotalStake, processedAccount.TotalStake)
	}
}

const (
	delegation = iota
	validator
	delegationLegacy
)

func makeMapFromArrays(keys []string, accounts []*data.AccountInfoWithStakeValues) map[string]*data.AccountInfoWithStakeValues {
	newMap := make(map[string]*data.AccountInfoWithStakeValues)
	for idx := 0; idx < len(keys); idx++ {
		newMap[keys[idx]] = accounts[idx]
	}
	return newMap
}

func generateAccounts(acctType int, numAccounts int) []*data.AccountInfoWithStakeValues {
	accts := make([]*data.AccountInfoWithStakeValues, 0)
	for idx := 0; idx < numAccounts; idx++ {
		acct := data.AccountInfoWithStakeValues{}

		switch acctType {
		case delegation:
			acct.Delegation = generateRandomBigIntString()
			acct.DelegationNum = core.ComputeBalanceAsFloat(acct.Delegation)
		case validator:
			acct.ValidatorsActive = generateRandomBigIntString()
			acct.ValidatorsActiveNum = core.ComputeBalanceAsFloat(acct.ValidatorsActive)
			acct.ValidatorTopUp = generateRandomBigIntString()
			acct.ValidatorTopUpNum = core.ComputeBalanceAsFloat(acct.Delegation)
		case delegationLegacy:
			acct.DelegationLegacyActive = generateRandomBigIntString()
			acct.DelegationLegacyActiveNum = core.ComputeBalanceAsFloat(acct.DelegationLegacyActive)
			acct.DelegationLegacyWaiting = generateRandomBigIntString()
			acct.DelegationLegacyWaitingNum = core.ComputeBalanceAsFloat(acct.DelegationLegacyWaiting)
		}

		accts = append(accts, &acct)
	}

	return accts
}

func generateRandomBigIntString() string {
	blk := make([]byte, 10)
	_, _ = rand.Read(blk)

	return big.NewInt(0).SetBytes(blk).String()
}
