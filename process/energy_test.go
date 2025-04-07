package process

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/config"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/mocks"
	"github.com/stretchr/testify/require"
)

func TestExtractAddressesAndEnergy(t *testing.T) {
	t.Parallel()

	pubKey, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	auth := core.FetchAuthenticationData(config.APIConfig{})
	accountsWithEnergyGetter, err := NewAccountsGetter(&mocks.RestClientStub{}, pubKey, auth, config.GeneralConfig{}, &mocks.ElasticClientStub{})
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
				EnergyDetails: &data.EnergyDetails{
					LastUpdateEpoch:   1891,
					Amount:            "5328000000000000000000",
					TotalLockedTokens: "32000000000000000000",
				},
			},
		},
		"erd1ejjwyzrdj053vcs5nhupxn6kha8audf4mla6tth9339zmcx52w5q7djae2": {
			StakeInfo: data.StakeInfo{
				Energy:    "273000000000000000000",
				EnergyNum: 273,
				EnergyDetails: &data.EnergyDetails{
					LastUpdateEpoch:   1891,
					Amount:            "4173000000000000000000",
					TotalLockedTokens: "25000000000000000000",
				},
			},
		},
		"erd1yhhzgv5ql3h8gppy5286grre23vfgw68tnth7dmcl8ywpd9puluqlcvvw9": {
			StakeInfo: data.StakeInfo{
				Energy:    "12625000000000000000000000",
				EnergyNum: 12625000,
				EnergyDetails: &data.EnergyDetails{
					LastUpdateEpoch:   1881,
					Amount:            "96455000000000000000000000",
					TotalLockedTokens: "505000000000000000000000",
				},
			}},
		"erd188lxgu4m889yht73t3svs4lxknfqtv2vgymgzz283x6wv4hw9nwq0cgw0v": {
			StakeInfo: data.StakeInfo{
				Energy:    "63371454581200312235",
				EnergyNum: 63.3714545812,
				EnergyDetails: &data.EnergyDetails{
					LastUpdateEpoch:   1881,
					Amount:            "4544871637244977820221",
					TotalLockedTokens: "26996989052191430771",
				},
			}},
	}, res)
}

func readJson(path string) string {
	jsonFile, _ := os.Open(path)
	byteValue, _ := ioutil.ReadAll(jsonFile)

	return string(byteValue)
}

func TestExtractBlockInfo(t *testing.T) {
	t.Parallel()

	response := `{"blockInfo":{"hash":"52a2e3c800d03b1499e3cbc57431ee5f122e1bf0e1065fa05578f2d58621f7a0","nonce":3576295,"rootHash":"1829f8c869318f1c5ddc8a887fc2bb206b42fa9a484b8beb94e7873b633cdc61"}}`

	blockInfo, err := extractBlockInfo([]byte(response))
	require.Nil(t, err)
	require.Equal(t, &data.BlockInfo{
		Hash:     "52a2e3c800d03b1499e3cbc57431ee5f122e1bf0e1065fa05578f2d58621f7a0",
		Nonce:    uint64(3576295),
		RootHash: "1829f8c869318f1c5ddc8a887fc2bb206b42fa9a484b8beb94e7873b633cdc61",
	}, blockInfo)
}
