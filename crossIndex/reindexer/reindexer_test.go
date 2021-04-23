package reindexer

import (
	"testing"

	dataIndexer "github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/elasticClient"
	"github.com/stretchr/testify/require"
)

func TestReindexer(t *testing.T) {
	sourceIndexer, err := elasticClient.NewElasticClient(
		data.EsClientConfig{
			Address: "https://internal-index.elrond.com",
		},
	)
	require.NoError(t, err)

	destinationIndexer, err := elasticClient.NewElasticClient(
		data.EsClientConfig{
			Address: "http://localhost:9200",
		},
	)
	require.NoError(t, err)

	reindexer, err := New(sourceIndexer, destinationIndexer)
	require.NoError(t, err)
	accounts := map[string]*data.AccountInfoWithStakeValues{
		"erd195fe57d7fm5h33585sc7wl8trqhrmy85z3dg6f6mqd0724ymljxq3zjemc": {
			AccountInfo: dataIndexer.AccountInfo{},
			StakeInfo: data.StakeInfo{
				DelegationLegacyWaiting: "75000000000000000000",
				ValidatorTopUp:          "999999999999999999999",
			},
		},
	}
	err = reindexer.ReindexAccounts("accounts", "accounts-000001_8", accounts)
	require.Nil(t, err)
}
