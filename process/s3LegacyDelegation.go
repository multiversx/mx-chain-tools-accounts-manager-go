package process

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/tidwall/gjson"
)

const (
	legacyDelegationShardID = uint32(2)
)

var pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

func (b *s3Balances) getFullActiveAndWaitingListAccounts(epoch uint32) (map[string]*data.AccountInfoWithStakeValues, error) {
	legacyDelegationStateBytes, err := b.sClient.GetFile(prepareFileName(legacyDelegationStateFilePrefix, legacyDelegationShardID, epoch))
	if err != nil {
		return nil, err
	}

	activeList, waitingList, err := parseLegacyDelegationState(legacyDelegationStateBytes)

	accountsMap := make(map[string]*data.AccountInfoWithStakeValues)
	for address, value := range activeList {
		_, found := accountsMap[address]
		if !found {
			accountsMap[address] = &data.AccountInfoWithStakeValues{
				StakeInfo: data.StakeInfo{
					DelegationLegacyActive:    value,
					DelegationLegacyActiveNum: core.ComputeBalanceAsFloat(value),
				},
			}

			continue
		}

		valueStake, valueStakeNum := computeTotalBalance(value, accountsMap[address].DelegationLegacyActive)

		accountsMap[address].DelegationLegacyActive = valueStake
		accountsMap[address].DelegationLegacyActiveNum = valueStakeNum
	}

	for key, value := range waitingList {
		_, ok := accountsMap[key]
		if !ok {
			accountsMap[key] = &data.AccountInfoWithStakeValues{
				StakeInfo: data.StakeInfo{
					DelegationLegacyWaiting:    value,
					DelegationLegacyWaitingNum: core.ComputeBalanceAsFloat(value),
				},
			}

			continue
		}

		valueWaiting, valueWaitingNum := computeTotalBalance(value, accountsMap[key].DelegationLegacyWaiting)

		accountsMap[key].DelegationLegacyWaiting = valueWaiting
		accountsMap[key].DelegationLegacyWaitingNum = valueWaitingNum
	}

	return accountsMap, nil
}

func parseLegacyDelegationState(state []byte) (map[string]string, map[string]string, error) {
	pairs := gjson.Get(string(state), "pairs")

	pairsMap := make(map[string]string)
	err := json.Unmarshal([]byte(pairs.String()), &pairsMap)
	if err != nil {
		log.Warn("cannot unmarshal accounts info", "error", err.Error())
		return nil, nil, err
	}

	userAddressID := make(map[string]int)
	for key, value := range pairsMap {
		keyDecoded, errD := hex.DecodeString(key)
		if errD != nil {
			return nil, nil, errD
		}
		const userAddressPrefix = "user_id"
		if bytes.HasPrefix(keyDecoded, []byte(userAddressPrefix)) && len(keyDecoded) == len(userAddressPrefix)+32 {
			userAddress, errE := pubKeyConverter.Encode(keyDecoded[len(userAddressPrefix):])
			if errE != nil {
				return nil, nil, errE
			}

			valueDecoded, errE := hex.DecodeString(value)
			if errE != nil {
				return nil, nil, errE
			}
			bigValue := big.NewInt(0).SetBytes(valueDecoded)
			userAddressID[userAddress] = int(bigValue.Int64())
		}
	}

	userIDTotalDelegationActive := make(map[int]*big.Int)
	userIDTotalDelegationWaiting := make(map[int]*big.Int)
	for key, value := range pairsMap {
		keyDecoded, errD := hex.DecodeString(key)
		if errD != nil {
			return nil, nil, errD
		}
		hasCorrectPrefix := bytes.HasPrefix(keyDecoded, []byte("f")) && !bytes.HasPrefix(keyDecoded, []byte("f_max_id")) &&
			!bytes.HasPrefix(keyDecoded, []byte("ftype")) && !bytes.HasPrefix(keyDecoded, []byte("fuser"))

		if hasCorrectPrefix {
			decodedValue, errE := hex.DecodeString(value)
			if errE != nil {
				return nil, nil, errE
			}

			if decodedValue[0] == byte(4) { // active
				userID := big.NewInt(0).SetBytes(decodedValue[1:5]).Int64()
				amountLen := big.NewInt(0).SetBytes(decodedValue[5:9]).Int64()
				amount := big.NewInt(0).SetBytes(decodedValue[9 : 9+amountLen])

				_, ok := userIDTotalDelegationActive[int(userID)]
				if !ok {
					userIDTotalDelegationActive[int(userID)] = amount
				} else {
					userIDTotalDelegationActive[int(userID)].Add(userIDTotalDelegationActive[int(userID)], amount)
				}
			} else if decodedValue[0] == byte(1) { // waiting
				// 1:9 creation date
				userID := big.NewInt(0).SetBytes(decodedValue[9:13]).Int64()
				amountLen := big.NewInt(0).SetBytes(decodedValue[13:17]).Int64()
				amount := big.NewInt(0).SetBytes(decodedValue[17 : 17+amountLen])
				_, ok := userIDTotalDelegationWaiting[int(userID)]
				if !ok {
					userIDTotalDelegationWaiting[int(userID)] = amount
				} else {
					userIDTotalDelegationWaiting[int(userID)].Add(userIDTotalDelegationWaiting[int(userID)], amount)
				}
			}
		}
	}

	addressActiveDelegation := make(map[string]string)
	addressWaitingDelegation := make(map[string]string)
	for address, userID := range userAddressID {
		delegationValue, ok := userIDTotalDelegationActive[userID]
		if !ok {
			//fmt.Printf("cannot find delegation value for address %s\n", address)
			//return fmt.Errorf("cannot find delegation value for address %s", address)
			continue
		}

		addressActiveDelegation[address] = delegationValue.String()
	}

	for address, userID := range userAddressID {
		delegationValue, ok := userIDTotalDelegationWaiting[userID]
		if !ok {
			//fmt.Printf("cannot find delegation value for address %s\n", address)
			//return fmt.Errorf("cannot find delegation value for address %s", address)
			continue
		}

		addressWaitingDelegation[address] = delegationValue.String()
	}

	return addressActiveDelegation, addressWaitingDelegation, nil
}
