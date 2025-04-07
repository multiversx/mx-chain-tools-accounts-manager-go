package process

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/config"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/mocks"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestReadDelegationLegacyStateFromFileAndExtractData(t *testing.T) {
	t.Parallel()

	pubKey, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	auth := core.FetchAuthenticationData(config.APIConfig{})
	accountsGetterLegacyDelegation, err := NewAccountsGetter(&mocks.RestClientStub{}, pubKey, auth, config.GeneralConfig{}, &mocks.ElasticClientStub{})
	require.Nil(t, err)

	testData := readJson("./testdata/delegation-legacy.json")

	pairs := gjson.Get(testData, "data.pairs")

	pairsMap := make(map[string]string)
	err = json.Unmarshal([]byte(pairs.String()), &pairsMap)
	require.Nil(t, err)

	res, err := accountsGetterLegacyDelegation.extractDelegationLegacyData(pairsMap)
	require.Nil(t, err)
	fmt.Println(len(res))

	// check random address
	info := res["erd1q5h0tjdkgl4pkn57qnljjgsamzvx548t5s02636wnynmtqmevv2q52lxdw"]
	require.Equal(t, "35000000000000000000", info.UnDelegateLegacy)

	info = res["erd13wanstz0wmjv0ashn2760cl2a2l5y6gwz2lay270347ujshm9unsvt73fn"]
	require.Equal(t, "323053985724926356758", info.DelegationLegacyActive)
}
