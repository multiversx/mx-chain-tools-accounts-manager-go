package convert

import (
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
)

const (
	numDecimalsInFloatBalance = 10
	denomination              = 18
)

var balancePrecision = math.Pow(10, float64(numDecimalsInFloatBalance))
var dividerForDenomination = math.Pow(10, float64(core.MaxInt(denomination, 0)))

// ComputeBalanceAsFloat will compute a string balance in float
func ComputeBalanceAsFloat(balance string) float64 {
	balanceBigInt, _ := big.NewInt(0).SetString(balance, 10)
	balanceBigFloat := big.NewFloat(0).SetInt(balanceBigInt)
	balanceFloat64, _ := balanceBigFloat.Float64()

	bal := balanceFloat64 / dividerForDenomination
	balanceFloatWithDecimals := math.Round(bal*balancePrecision) / balancePrecision

	return core.MaxFloat64(balanceFloatWithDecimals, 0)
}
