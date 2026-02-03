package process

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/sharding"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/tidwall/gjson"
)

const (
	hexEncodedEnergyPrefix = "75736572456e65726779"
	pathIterateKeys        = "/address/iterate-keys"
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

	pairs := gjson.Get(string(genericAPIResponse.Data), "pairs")

	keyValueMap := make(map[string]string)
	err = json.Unmarshal([]byte(pairs.String()), &keyValueMap)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot unmarshal account storage, error: %s", err.Error())
	}

	accountsWithEnergy, err := ag.extractAddressesAndEnergy(keyValueMap, currentEpoch)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot extract accounts with energy %s", err.Error())
	}

	blockInfo, err := extractBlockInfo(genericAPIResponse.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot extract block info %s", err.Error())
	}

	return accountsWithEnergy, blockInfo, nil
}

// GetAccountsWithEnergyV2 wii return the accounts with energy using the endpoint /address/iterate-keys
func (ag *accountsGetter) GetAccountsWithEnergyV2(currentEpoch uint32) (map[string]*data.AccountInfoWithStakeValues, *data.BlockInfo, error) {
	if ag.energyContractAddress == "" {
		return map[string]*data.AccountInfoWithStakeValues{}, nil, nil
	}

	defer logExecutionTime(time.Now(), "Fetched accounts from energy contract v2")

	lastestBlockNonce, err := ag.getLatestBlockNonceForShardAddress(ag.energyContractAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get latest block nonce for address %s: %w", ag.energyContractAddress, err)
	}

	log.Info("block nonce for shard", "nonce", lastestBlockNonce)

	path := fmt.Sprintf("%s?blockNonce=%d", pathIterateKeys, lastestBlockNonce)
	request := &data.IterateKeysRequest{
		Address:       ag.energyContractAddress,
		NumKeys:       15_000,
		IteratorState: nil,
	}

	response := &data.IterateKeysAPIResponse{}
	keyValueMap := make(map[string]string)
	count := 0
	for {
		err = ag.restClient.CallPostRestEndPoint(path, request, response, core.GetEmptyApiCredentials())
		if err != nil {
			return nil, nil, err
		}

		if response.Error != "" {
			return nil, nil, fmt.Errorf("cannot iterate keys energy contract %s", response.Error)
		}

		for key, value := range response.Data.Pairs {
			keyValueMap[key] = value
		}

		log.Info("fetched keys", "idx", count, "total pairs", len(keyValueMap))
		count++

		if len(response.Data.NewIteratorState) > 0 {
			request.IteratorState = response.Data.NewIteratorState
			continue
		}
		break
	}

	accountsWithEnergy, err := ag.extractAddressesAndEnergy(keyValueMap, currentEpoch)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot extract accounts with energy %s", err.Error())
	}

	return accountsWithEnergy, response.Data.BlockInfo, nil
}

func (ag *accountsGetter) getLatestBlockNonceForShardAddress(address string) (uint64, error) {
	decodedKey, err := ag.pubKeyConverter.Decode(address)
	if err != nil {
		return 0, fmt.Errorf("cannot get latest block nonce: %s", err.Error())
	}

	shardID := sharding.ComputeShardID(decodedKey, 3)

	genericAPIResponse := &data.GenericAPIResponse{}
	err = ag.restClient.CallGetRestEndPoint(fmt.Sprintf("/network/status/%d", shardID), genericAPIResponse, core.GetEmptyApiCredentials())
	if err != nil {
		return 0, err
	}
	if genericAPIResponse.Error != "" {
		return 0, fmt.Errorf("cannot get latest block nonce for shard %d: network status error: %s", shardID, genericAPIResponse.Error)
	}

	nonce := gjson.Get(string(genericAPIResponse.Data), "status.erd_nonce")
	return uint64(nonce.Num), nil

}

func (ag *accountsGetter) extractAddressesAndEnergy(keyValueMap map[string]string, currentEpoch uint32) (map[string]*data.AccountInfoWithStakeValues, error) {
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

	return ag.pubKeyConverter.SilentEncode(addressBytes, log), true
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
