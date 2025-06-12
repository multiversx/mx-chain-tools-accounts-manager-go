package process

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/tidwall/gjson"
)

const (
	hexEncodedEnergyPrefix = "75736572456e65726779"
)

// GetAccountsWithEnergy will return accounts with energy
func (ag *accountsGetter) GetAccountsWithEnergy(currentEpoch uint32) (map[string]*data.AccountInfoWithStakeValues, *data.BlockInfo, error) {
	if ag.energyContractAddress == "" {
		return map[string]*data.AccountInfoWithStakeValues{}, nil, nil
	}

	defer logExecutionTime(time.Now(), "Fetched accounts from energy contract")

	genericAPIResponse := &data.GenericAPIResponse{}
	path := fmt.Sprintf(pathAccountKeys, ag.energyContractAddress)
	err := ag.restClient.CallGetRestEndPoint(path, genericAPIResponse, core.GetEmptyApiCredentials())
	if err != nil {
		return nil, nil, err
	}
	if genericAPIResponse.Error != "" {
		return nil, nil, fmt.Errorf("cannot get accounts with energy %s", genericAPIResponse.Error)
	}

	accountsWithEnergy, err := ag.extractAddressesAndEnergy(genericAPIResponse.Data, currentEpoch)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot extract accounts with energy %s", err.Error())
	}

	blockInfo, err := extractBlockInfo(genericAPIResponse.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot extract block info %s", err.Error())
	}

	return accountsWithEnergy, blockInfo, nil
}

func (ag *accountsGetter) extractAddressesAndEnergy(accountStorage []byte, currentEpoch uint32) (map[string]*data.AccountInfoWithStakeValues, error) {
	pairs := gjson.Get(string(accountStorage), "pairs")

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
		energyDetails, ok := extractEnergyFromValue(value)
		if !ok {
			continue
		}

		energyValue := calculateEnergyValueBasedOnCurrentEpoch(energyDetails, currentEpoch)
		energyNum := core.ComputeBalanceAsFloat(energyValue.String())

		accountsWithEnergy[address] = &data.AccountInfoWithStakeValues{
			StakeInfo: data.StakeInfo{
				Energy:        energyValue.String(),
				EnergyNum:     energyNum,
				EnergyDetails: energyDetails,
			},
		}
	}

	log.Info("accounts with energy", "num", len(accountsWithEnergy))

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

const (
	numBytesForBigValueLength = 4
	numBytesForU64Value       = 8
)

func extractEnergyFromValue(value string) (*data.EnergyDetails, bool) {
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
	if len(amountValueInBytes) > 0 && amountValueInBytes[0]&0x80 != 0 { // check MSB
		// Create 2^N where N is number of bits
		bitLen := len(amountValueInBytes) * 8
		twoPow := new(big.Int).Lsh(big.NewInt(1), uint(bitLen))
		amount.Sub(amount, twoPow)
	}

	totalLockedTokens := big.NewInt(0).SetBytes(lockTokenValueInBytes)
	energy := &data.EnergyDetails{
		Amount:            amount.String(),
		LastUpdateEpoch:   uint32(lastUpdateEpochUint64),
		TotalLockedTokens: totalLockedTokens.String(),
	}

	return energy, true
}

func calculateEnergyValueBasedOnCurrentEpoch(energy *data.EnergyDetails, currentEpoch uint32) *big.Int {
	coefficient := currentEpoch - energy.LastUpdateEpoch
	totalLockedTokens, ok := big.NewInt(0).SetString(energy.TotalLockedTokens, 10)
	if !ok {
		totalLockedTokens = big.NewInt(0)
	}
	amount, ok := big.NewInt(0).SetString(energy.Amount, 10)
	if !ok {
		amount = big.NewInt(0)
	}

	valueToSubtract := big.NewInt(0).Mul(big.NewInt(int64(coefficient)), totalLockedTokens)
	energyValue := big.NewInt(0).Sub(amount, valueToSubtract)

	log.Trace(
		"calculateEnergyValueBasedOnCurrentEpoch",
		"current epoch", currentEpoch,
		"last update epoch", energy.LastUpdateEpoch,
		"amount", core.ComputeBalanceAsFloat(energy.Amount),
		"total locked tokens", core.ComputeBalanceAsFloat(energy.TotalLockedTokens),
		"energy", core.ComputeBalanceAsFloat(energyValue.String()))

	return energyValue
}

func extractBlockInfo(responseWithBlockInfo []byte) (*data.BlockInfo, error) {
	blockInfoData := gjson.Get(string(responseWithBlockInfo), "blockInfo")

	blockInfo := &data.BlockInfo{}
	err := json.Unmarshal([]byte(blockInfoData.String()), &blockInfo)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal account storage, error: %s", err.Error())
	}

	return blockInfo, nil
}
