package core

import (
	"math/big"

	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

// MergeElasticAndRestAccounts will merge additional data from the rest into the existing data from elastic search
func MergeElasticAndRestAccounts(
	accountsES, accountsRest map[string]*data.AccountInfoWithStakeValues,
) map[string]*data.AccountInfoWithStakeValues {
	accounts := make(map[string]*data.AccountInfoWithStakeValues)

	for address, account := range accountsES {
		accounts[address] = account

		accountRest, ok := accountsRest[address]
		if !ok {
			accounts[address].TotalBalanceWithStake = accounts[address].Balance
			accounts[address].TotalBalanceWithStakeNum = accounts[address].BalanceNum

			continue
		}

		accounts[address].StakeInfo = accountRest.StakeInfo

		totalBalanceWithStake, totalBalanceWithStakeNum := computeTotalBalance(
			accounts[address].Balance,
			accountRest.TotalStake,
			accountRest.TotalUnDelegate,
		)

		accounts[address].TotalBalanceWithStake = totalBalanceWithStake
		accounts[address].TotalBalanceWithStakeNum = totalBalanceWithStakeNum
	}

	return accounts
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
		totalBalanceFloat += ComputeBalanceAsFloat(balance)
	}

	return totalBalance.String(), totalBalanceFloat
}
