package process

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/config"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/mocks"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestAccountsGetter_ValidatorsAccountsPutUnDelegatedValues(t *testing.T) {
	t.Parallel()

	pubKeyConverter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)

	ag, err := NewAccountsGetter(&mocks.RestClientStub{
		CallPostRestEndPointCalled: func(path string, dataD interface{}, response interface{}, authenticationData data.RestApiAuthenticationData) error {
			responseVmValue := response.(*data.ResponseVmValue)
			responseVmValue.Data = data.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnData: [][]byte{big.NewInt(1000000000000000000).Bytes(), []byte("")},
					ReturnCode: vmcommon.Ok.String(),
				},
			}

			return nil
		},
	}, pubKeyConverter, data.RestApiAuthenticationData{}, config.GeneralConfig{
		ValidatorsContract: "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l",
	}, &mocks.ElasticClientStub{})
	require.Nil(t, err)

	accountsWithStakeJson := readJson("./testdata/accounts-with-stake.json")
	accountsWithStake := make(map[string]*data.AccountInfoWithStakeValues)
	err = json.Unmarshal([]byte(accountsWithStakeJson), &accountsWithStake)
	require.Nil(t, err)

	err = ag.putUndelegatedValuesFromValidatorsContract(accountsWithStake)
	require.Nil(t, err)

	for _, account := range accountsWithStake {
		require.Equal(t, account.UnDelegateValidator, "1000000000000000000")
		require.Equal(t, account.UnDelegateValidatorNum, float64(1))
	}
}
