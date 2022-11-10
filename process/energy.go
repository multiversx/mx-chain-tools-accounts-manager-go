package process

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-accounts-manager/core"
	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/tidwall/gjson"
)

const (
	hexEncodedEnergyPrefix = "75736572456e65726779"
)

// GetAccountsWithEnergy will return accounts with energy
func (ag *accountsGetter) GetAccountsWithEnergy(currentEpoch uint32) (map[string]*data.AccountInfoWithStakeValues, error) {
	genericAPIResponse := &data.GenericAPIResponse{}
	path := fmt.Sprintf(pathAccountKeys, ag.energyContractAddress)
	err := ag.restClient.CallGetRestEndPoint(path, genericAPIResponse, core.GetEmptyApiCredentials())
	if err != nil {
		return nil, err
	}
	if genericAPIResponse.Error != "" {
		return nil, fmt.Errorf("get accounts with energy %s", genericAPIResponse.Error)
	}

	return ag.extractAddressesAndEnergy(genericAPIResponse.Data, currentEpoch)
}

func (ag *accountsGetter) extractAddressesAndEnergy(accountStorage []byte, currentEpoch uint32) (map[string]*data.AccountInfoWithStakeValues, error) {
	pairs := gjson.Get(string(accountStorage), "data.pairs")

	keyValueMap := make(map[string]string)
	err := json.Unmarshal([]byte(pairs.String()), &keyValueMap)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal account storage, error: %s", err.Error())
	}

	accountsWithEnergy := make(map[string]*data.AccountInfoWithStakeValues)
	for key, value := range keyValueMap {
		address, ok := ag.extractAddressFromKey(key)
		if !ok {
			continue
		}
		energy, ok := ag.extractEnergyFromValue(value)
		if !ok {
			continue
		}

		energyValue := calculateEnergyValueBasedOnCurrentEpoch(energy, currentEpoch)

		// ignore address with energyValue less then zero
		zero := big.NewInt(0)
		if zero.Cmp(energyValue) > 0 {
			continue
		}

		accountsWithEnergy[address] = &data.AccountInfoWithStakeValues{
			StakeInfo: data.StakeInfo{
				Energy:    energyValue.String(),
				EnergyNum: core.ComputeBalanceAsFloat(energyValue.String()),
			},
		}
	}

	return accountsWithEnergy, nil
}

func (ag *accountsGetter) extractAddressFromKey(key string) (string, bool) {
	hasPrefix := strings.HasPrefix(key, hexEncodedEnergyPrefix)
	if !hasPrefix {
		return "", false
	}

	hexEncodedAddress := strings.ReplaceAll(key, hexEncodedEnergyPrefix, "")
	addressBytes, err := hex.DecodeString(hexEncodedAddress)
	if err != nil {
		log.Warn("cannot decode address from key", "error", err)
		return "", false
	}

	return ag.pubKeyConverter.Encode(addressBytes), true
}

type energyStruc struct {
	Amount            *big.Int
	LastUpdateEpoch   uint32
	TotalLockedTokens *big.Int
}

const (
	numBytesForBigValueLength = 4
	numBytesForU64Value       = 8
)

func (ag *accountsGetter) extractEnergyFromValue(value string) (*energyStruc, bool) {
	decodedBytes, err := hex.DecodeString(value)
	if err != nil {
		log.Warn("cannot decode energy structure bytes", "error", err)
		return nil, false
	}

	// decodedBytes contains
	// -----------------------------------------------------------------------
	// |l11|l12|l13|l14|a1|a2|..|ax|e1|e2|..|e8|l2|l22|l23|l24|lt1|lt2|..|ltx|
	// -----------------------------------------------------------------------
	// [l11,l14] --- length of Amount
	// [a1,ax] --- amount bytes
	// [e1,e8] -- last_update_epoch
	// [l21,l24] -- length of LockedTokens
	// [lt1,ltx] --- total_locked_tokens bytes

	///// extract amount ////////////////////////////////////////////////////////////
	bigIntLengthBytes := decodedBytes[0:numBytesForBigValueLength]
	amountValueNumBytes := binary.BigEndian.Uint32(bigIntLengthBytes)
	amountValueEndIndex := numBytesForBigValueLength + amountValueNumBytes
	amountValueInBytes := decodedBytes[numBytesForBigValueLength:amountValueEndIndex]
	/////////////////////////////////////////////////////////////////////////////////

	///// extract last_update_epoch /////////////////////////////////////////////////
	lastUpdateEpochEndIndex := amountValueEndIndex + numBytesForU64Value
	lastUpdateEpochBytes := decodedBytes[amountValueEndIndex:lastUpdateEpochEndIndex]
	lastUpdateEpochUint64 := binary.BigEndian.Uint64(lastUpdateEpochBytes)
	/////////////////////////////////////////////////////////////////////////////////

	///// extract total locked tokens ///////////////////////////////////////////////
	secondBigIntLengthEndIndex := lastUpdateEpochEndIndex + numBytesForBigValueLength
	secondBigIntLength := decodedBytes[lastUpdateEpochEndIndex:secondBigIntLengthEndIndex]
	lockTokenNumBytes := binary.BigEndian.Uint32(secondBigIntLength)
	lockTokenValueInBytes := decodedBytes[secondBigIntLengthEndIndex : secondBigIntLengthEndIndex+lockTokenNumBytes]
	/////////////////////////////////////////////////////////////////////////////////

	amount := big.NewInt(0).SetBytes(amountValueInBytes)
	totalLockedTokens := big.NewInt(0).SetBytes(lockTokenValueInBytes)
	energy := &energyStruc{
		Amount:            amount,
		LastUpdateEpoch:   uint32(lastUpdateEpochUint64),
		TotalLockedTokens: totalLockedTokens,
	}

	return energy, true
}

func calculateEnergyValueBasedOnCurrentEpoch(energy *energyStruc, currentEpoch uint32) *big.Int {
	coefficient := currentEpoch - energy.LastUpdateEpoch
	valueToSubtract := big.NewInt(0).Mul(big.NewInt(int64(coefficient)), energy.TotalLockedTokens)
	energyValue := big.NewInt(0).Sub(energy.Amount, valueToSubtract)

	log.Trace(
		"calculateEnergyValueBasedOnCurrentEpoch",
		"current epoch", currentEpoch,
		"last update epoch", energy.LastUpdateEpoch,
		"amount", core.ComputeBalanceAsFloat(energy.Amount.String()),
		"total locked tokens", core.ComputeBalanceAsFloat(energy.TotalLockedTokens.String()),
		"energy", core.ComputeBalanceAsFloat(energyValue.String()))

	return energyValue
}