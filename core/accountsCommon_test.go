package core

import (
	"testing"

	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/stretchr/testify/require"
)

const zeros = "000000000000000000"

func TestAccountsProcessor_PrepareAccountsForReindexing(t *testing.T) {
	t.Parallel()

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

	mReturn := MergeElasticAndRestAccounts(mES, mR)
	require.Len(t, mReturn, 2)

	for _, addr := range addresses {
		require.Equal(t, mES[addr].Balance, mReturn[addr].Balance)
		require.Equal(t, mES[addr].BalanceNum, mReturn[addr].BalanceNum)

		require.Equal(t, mES[addr].DelegationNum, mReturn[addr].DelegationNum)
		require.Equal(t, mES[addr].Delegation, mReturn[addr].Delegation)
	}
}
