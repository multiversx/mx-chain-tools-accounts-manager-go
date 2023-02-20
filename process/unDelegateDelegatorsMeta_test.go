package process

import (
	"encoding/json"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/config"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/mocks"
	"github.com/stretchr/testify/require"
)

func TestAccountsGetter_DelegationMetaPutUnDelegatedValues(t *testing.T) {
	pubKeyConverter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)

	ag, err := NewAccountsGetter(&mocks.RestClientStub{}, pubKeyConverter, data.RestApiAuthenticationData{}, config.GeneralConfig{}, &mocks.ElasticClientStub{
		DoScrollRequestAllDocumentsCalled: func(index string, body []byte, handlerFunc func(responseBytes []byte) error) error {
			// read from a file delegators
			return handlerFunc(nil)
		},
	})
	require.Nil(t, err)

	accountsWithStakeJson := readJson("./testdata/accounts-with-stake.json")
	accountsWithStake := make(map[string]*data.AccountInfoWithStakeValues)
	err = json.Unmarshal([]byte(accountsWithStakeJson), &accountsWithStake)
	require.Nil(t, err)

	err = ag.unDelegatedInfoProc.putUnDelegateInfoFromStakingProviders(accountsWithStake)
	require.Nil(t, err)

	//for _, account := range accountsWithStake {
	//	require.Equal(t, account.UnDelegateLegacy, "2000000000000000000")
	//	require.Equal(t, account.UnDelegateLegacyNum, float64(2))
	//}
}
