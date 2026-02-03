package mocks

import (
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

// AccountsGetterStub -
type AccountsGetterStub struct {
	GetLegacyDelegatorsAccountsCalled func() (map[string]*data.AccountInfoWithStakeValues, error)
	GetValidatorsAccountsCalled       func() (map[string]*data.AccountInfoWithStakeValues, error)
	GetDelegatorsAccountsCalled       func() (map[string]*data.AccountInfoWithStakeValues, error)
}

// GetAccountsWithEnergy -
func (a *AccountsGetterStub) GetAccountsWithEnergy(_ uint32) (map[string]*data.AccountInfoWithStakeValues, *data.BlockInfo, error) {
	return nil, nil, nil
}

// GetAccountsWithEnergyV2 -
func (a *AccountsGetterStub) GetAccountsWithEnergyV2(_ uint32) (map[string]*data.AccountInfoWithStakeValues, *data.BlockInfo, error) {
	return nil, nil, nil
}

// GetLKMEXStakeAccounts -
func (a *AccountsGetterStub) GetLKMEXStakeAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	return nil, nil
}

// GetLegacyDelegatorsAccounts -
func (a *AccountsGetterStub) GetLegacyDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	if a.GetLegacyDelegatorsAccountsCalled != nil {
		return a.GetLegacyDelegatorsAccountsCalled()
	}
	return nil, nil
}

// GetValidatorsAccounts -
func (a *AccountsGetterStub) GetValidatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	if a.GetValidatorsAccountsCalled != nil {
		return a.GetValidatorsAccountsCalled()
	}
	return nil, nil
}

// GetDelegatorsAccounts -
func (a *AccountsGetterStub) GetDelegatorsAccounts() (map[string]*data.AccountInfoWithStakeValues, error) {
	if a.GetDelegatorsAccountsCalled != nil {
		return a.GetDelegatorsAccountsCalled()
	}
	return nil, nil
}
