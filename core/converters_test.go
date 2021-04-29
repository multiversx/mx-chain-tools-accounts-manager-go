package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeBalanceAsFloat(t *testing.T) {
	t.Parallel()

	require.Equal(t, float64(0), ComputeBalanceAsFloat(""))
	require.Equal(t, float64(0), ComputeBalanceAsFloat("aaaaa"))
	require.Equal(t, 1.5, ComputeBalanceAsFloat("1500000000000000000"))
}
