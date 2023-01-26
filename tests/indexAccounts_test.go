package tests

import (
	"fmt"
	"math/rand"
	"testing"

	dataIndexer "github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/elasticClient"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/process"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/process/accountsIndexer"
	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	t.Skip("ignore this test")

	cfg := data.EsClientConfig{
		Address:  "http://localhost:9200",
		Username: "",
		Password: "",
	}
	ec, err := elasticClient.NewElasticClient(cfg)
	require.Nil(t, err)

	numberOfAccounts := 1

	for idx := 0; idx < 1; idx++ {
		generateAccountsAndIndex(t, numberOfAccounts, ec)
	}
}

func generateAccountsAndIndex(t *testing.T, numberOfAccounts int, handler process.ElasticClientHandler) {
	ap, _ := accountsIndexer.NewAccountsIndexer(handler)
	err := ap.IndexAccounts(generateAccounts(numberOfAccounts), "accounts-000001")
	require.Nil(t, err)

	fmt.Println("DONE")
}

func generateAccounts(num int) map[string]*data.AccountInfoWithStakeValues {
	accs := make(map[string]*data.AccountInfoWithStakeValues, num)
	for idx := 0; idx < num; idx++ {
		accs[randStringRunes(32)] = generateAccount()
	}

	return accs
}

func generateAccount() *data.AccountInfoWithStakeValues {
	return &data.AccountInfoWithStakeValues{
		AccountInfo: dataIndexer.AccountInfo{
			Address:         randStringRunes(32),
			Nonce:           0,
			Balance:         randStringRunes(18),
			BalanceNum:      0,
			TokenIdentifier: randStringRunes(6),
			Properties:      randStringRunes(5),
			IsSender:        false,
		},
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789<>/;';'!@#$%^&")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
