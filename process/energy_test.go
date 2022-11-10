package process

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-accounts-manager/config"
	"github.com/ElrondNetwork/elrond-accounts-manager/core"
	"github.com/ElrondNetwork/elrond-accounts-manager/mocks"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/stretchr/testify/require"
)

func TestExtractAddressesAndEnergy(t *testing.T) {
	t.Parallel()

	pubKey, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	auth := core.FetchAuthenticationData(config.APIConfig{})
	accountsWithEnergyGetter, err := NewAccountsGetter(&mocks.RestClientStub{}, pubKey, auth, config.GeneralConfig{})
	require.Nil(t, err)

	testData := readJson("./testdata/account-storage.json")
	res, err := accountsWithEnergyGetter.extractAddressesAndEnergy([]byte(testData), 2047)
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Len(t, res, 4)
}

func readJson(path string) string {
	jsonFile, _ := os.Open(path)
	byteValue, _ := ioutil.ReadAll(jsonFile)

	return string(byteValue)
}
