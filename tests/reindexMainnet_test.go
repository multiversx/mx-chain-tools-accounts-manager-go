package tests

import (
	"testing"

	"github.com/ElrondNetwork/elrond-accounts-manager/tests/accountsReindex"
	"github.com/stretchr/testify/require"
)

func TestReindexMainnnet(t *testing.T) {
	client, _ := accountsReindex.NewElasticSearchClient(
		"https://internal-index.elrond.com",
		"https://search-testing-mihai-cupm4ru4fsbpsgkikuqx6oexie.eu-central-1.es.amazonaws.com")
	err := client.ProcessAllAccounts()
	require.NoError(t, err)
}
