package core

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-accounts-manager/data"
)

// MergeElasticAndRestAccounts will merge additional data from the rest into the existing data from elastic search
func MergeElasticAndRestAccounts(
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
