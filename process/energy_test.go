package process

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-accounts-manager/config"
	"github.com/ElrondNetwork/elrond-accounts-manager/core"
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
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
	require.Equal(t, map[string]*data.AccountInfoWithStakeValues{
		"erd10f7nnvqk8xvyd50f2sc5p4e0ru4alf99p3v7zfe4uvenra2esges39a9x7": {
			StakeInfo: data.StakeInfo{
				Energy:    "336000000000000000000",
				EnergyNum: 336,
			},
		},
		"erd1ejjwyzrdj053vcs5nhupxn6kha8audf4mla6tth9339zmcx52w5q7djae2": {
			StakeInfo: data.StakeInfo{
				Energy:    "273000000000000000000",
				EnergyNum: 273,
			},
		},
		"erd1yhhzgv5ql3h8gppy5286grre23vfgw68tnth7dmcl8ywpd9puluqlcvvw9": {
			StakeInfo: data.StakeInfo{
				Energy:    "12625000000000000000000000",
				EnergyNum: 12625000,
			}},
		"erd188lxgu4m889yht73t3svs4lxknfqtv2vgymgzz283x6wv4hw9nwq0cgw0v": {
			StakeInfo: data.StakeInfo{
				Energy:    "63371454581200312235",
				EnergyNum: 63.3714545812,
			}},
	}, res)
}

func readJson(path string) string {
	jsonFile, _ := os.Open(path)
	byteValue, _ := ioutil.ReadAll(jsonFile)

	return string(byteValue)
}
