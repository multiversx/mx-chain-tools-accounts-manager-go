package mocks

import (
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

type AccountsGetterStub struct {
	GetLegacyDelegatorsAccountsCalled func() (map[string]*data.AccountInfoWithStakeValues, error)
	GetValidatorsAccountsCalled       func() (map[string]*data.AccountInfoWithStakeValues, error)
	GetDelegatorsAccountsCalled       func() (map[string]*data.AccountInfoWithStakeValues, error)
}

func (a *AccountsGetterStub) GetAccountsWithEnergy(_ uint32) (map[string]*data.AccountInfoWithStakeValues, *data.BlockInfo, error) {
	return nil, nil, nil
}

func (a *AccountsGetterStub) GetLKMEXStakeAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	return nil, nil
}

func (a *AccountsGetterStub) GetLegacyDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	if a.GetLegacyDelegatorsAccountsCalled != nil {
		return a.GetLegacyDelegatorsAccountsCalled()
	}
	return nil, nil
}

func (a *AccountsGetterStub) GetValidatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	if a.GetValidatorsAccountsCalled != nil {
		return a.GetValidatorsAccountsCalled()
	}
	return nil, nil
}

func (a *AccountsGetterStub) GetDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	if a.GetDelegatorsAccountsCalled != nil {
		return a.GetDelegatorsAccountsCalled()
	}
	return nil, nil
}
