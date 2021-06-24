package accountsReindex

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const newIndexName = "accounts-cloned"

func TestCrossIndexer(t *testing.T) {
	crossIdx, err := NewCrossIndexer(
		"https://internal-index.elrond.com",
		"http://localhost:9200",
		newIndexName,
	)
	require.Nil(t, err)

	err = crossIdx.ReindexAccountsIndex()
	require.Nil(t, err)
}
