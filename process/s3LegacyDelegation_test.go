package process

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDelegationLegacy(t *testing.T) {
	legacyBytes, err := os.ReadFile("legacy_delegation_state_2_500")
	require.Nil(t, err)

	myArray := []byte("123456")

	fmt.Println(myArray[1:5])

	_, _, err = parseLegacyDelegationState(legacyBytes)
	require.Nil(t, err)
}
