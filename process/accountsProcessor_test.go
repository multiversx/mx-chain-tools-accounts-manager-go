package process

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-accounts-manager/convert"
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/mocks"
	"github.com/stretchr/testify/require"
)

const (
	zeros = "000000000000000000"
)

func TestAccountsProcessor_PrepareAccountsForReindexing(t *testing.T) {
	t.Parallel()

	ap, err := NewAccountsProcessor(&mocks.RestClientStub{}, &mocks.AccountsGetterStub{})
	require.Nil(t, err)

	accES1 := &data.AccountInfoWithStakeValues{}
	accES1.Balance = "1" + zeros
	accES1.BalanceNum = 1

	accES2 := &data.AccountInfoWithStakeValues{}
	accES1.Balance = "2" + zeros
	accES2.BalanceNum = 2

	accR1 := &data.AccountInfoWithStakeValues{}
	accR1.Delegation = "3" + zeros
	accR1.DelegationNum = 3

	accR2 := &data.AccountInfoWithStakeValues{}
	accR1.Delegation = "4" + zeros
	accR2.DelegationNum = 4

	addresses := []string{"1", "2"}

	mES := map[string]*data.AccountInfoWithStakeValues{
		addresses[0]: accES1,
		addresses[1]: accES2,
	}

	mR := map[string]*data.AccountInfoWithStakeValues{
		addresses[0]: accR1,
		addresses[1]: accR2,
	}

	mReturn := ap.PrepareAccountsForReindexing(mES, mR)
	require.Len(t, mReturn, 2)

	for _, addr := range addresses {
		require.Equal(t, mES[addr].Balance, mReturn[addr].Balance)
		require.Equal(t, mES[addr].BalanceNum, mReturn[addr].BalanceNum)

		require.Equal(t, mES[addr].DelegationNum, mReturn[addr].DelegationNum)
		require.Equal(t, mES[addr].Delegation, mReturn[addr].Delegation)
	}
}

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

	allAccounts, allAddresses, err := ap.GetAllAccountsWithStake()
	require.Nil(t, err)
	require.Equal(t, len(allAccounts), len(allAddresses))

	for addr, processedAccount := range allAccounts {
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
			acct.DelegationNum = convert.ComputeBalanceAsFloat(acct.Delegation)
		case validator:
			acct.ValidatorsActive = generateRandomBigIntString()
			acct.ValidatorsActiveNum = convert.ComputeBalanceAsFloat(acct.ValidatorsActive)
			acct.ValidatorTopUp = generateRandomBigIntString()
			acct.ValidatorTopUpNum = convert.ComputeBalanceAsFloat(acct.Delegation)
		case delegationLegacy:
			acct.DelegationLegacyActive = generateRandomBigIntString()
			acct.DelegationLegacyActiveNum = convert.ComputeBalanceAsFloat(acct.DelegationLegacyActive)
			acct.DelegationLegacyWaiting = generateRandomBigIntString()
			acct.DelegationLegacyWaitingNum = convert.ComputeBalanceAsFloat(acct.DelegationLegacyWaiting)
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
