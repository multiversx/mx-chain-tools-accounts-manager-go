package process

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

const (
	pathAddressKeys = "/address/%s/keys"

	userAddressPrefix = "user_id"

	addressLength = 32
)

func (ag *accountsGetter) extractDelegationLegacyData(
	pairsMap map[string]string,
) (map[string]*data.AccountInfoWithStakeValues, error) {
	userAddressID, err := ag.extractUsersIDMap(pairsMap)
	if err != nil {
		return nil, err
	}

	userIDTotalDelegationActive := make(map[int]*big.Int)
	userIDTotalDelegationWaiting := make(map[int]*big.Int)
	userIDTotalUnStaked := make(map[int]*big.Int)

	for key, value := range pairsMap {
		if !isCorrectPrefix(key) {
			continue
		}

		decodedValue, errI := hex.DecodeString(value)
		if errI != nil {
			return nil, errI
		}

		errI = ag.processDecodedValue(decodedValue, userIDTotalDelegationActive, userIDTotalDelegationWaiting, userIDTotalUnStaked)
		if errI != nil {
			return nil, errI
		}
	}

	mapAddressWithStake := ag.buildAddressWithStakeMap(userAddressID, userIDTotalDelegationActive, userIDTotalDelegationWaiting, userIDTotalUnStaked)

	log.Info("legacy delegators accounts", "num", len(mapAddressWithStake))

	return mapAddressWithStake, nil
}

func isCorrectPrefix(key string) bool {
	keyDecoded, err := hex.DecodeString(key)
	if err != nil {
		return false
	}
	return bytes.HasPrefix(keyDecoded, []byte("f")) &&
		!bytes.HasPrefix(keyDecoded, []byte("f_max_id")) &&
		!bytes.HasPrefix(keyDecoded, []byte("ftype")) &&
		!bytes.HasPrefix(keyDecoded, []byte("fuser"))
}

const (
	delegationActive   = byte(4)
	delegationWaiting  = byte(1)
	delegationUnStaked = byte(5)
)

func (ag *accountsGetter) processDecodedValue(decodedValue []byte, userIDTotalDelegationActive, userIDTotalDelegationWaiting, userIDTotalUnStaked map[int]*big.Int) error {
	var userID, amountLen int64
	var amount *big.Int

	switch decodedValue[0] {
	case delegationActive:
		userID = big.NewInt(0).SetBytes(decodedValue[1:5]).Int64()
		amountLen = big.NewInt(0).SetBytes(decodedValue[5:9]).Int64()
		amount = big.NewInt(0).SetBytes(decodedValue[9 : 9+amountLen])
		addToTotal(userIDTotalDelegationActive, int(userID), amount)

	case delegationWaiting:
		userID = big.NewInt(0).SetBytes(decodedValue[9:13]).Int64()
		amountLen = big.NewInt(0).SetBytes(decodedValue[13:17]).Int64()
		amount = big.NewInt(0).SetBytes(decodedValue[17 : 17+amountLen])
		addToTotal(userIDTotalDelegationWaiting, int(userID), amount)

	case delegationUnStaked:
		userID = big.NewInt(0).SetBytes(decodedValue[9:13]).Int64()
		amountLen = big.NewInt(0).SetBytes(decodedValue[13:17]).Int64()
		amount = big.NewInt(0).SetBytes(decodedValue[17 : 17+amountLen])
		addToTotal(userIDTotalUnStaked, int(userID), amount)
	}

	return nil
}

func addToTotal(userMap map[int]*big.Int, userID int, amount *big.Int) {
	if _, ok := userMap[userID]; !ok {
		userMap[userID] = amount
	} else {
		userMap[userID].Add(userMap[userID], amount)
	}
}

func (ag *accountsGetter) buildAddressWithStakeMap(
	userAddressID map[string]int,
	userIDTotalDelegationActive,
	userIDTotalDelegationWaiting,
	userIDTotalUnStaked map[int]*big.Int,
) map[string]*data.AccountInfoWithStakeValues {
	mapAddressWithStake := make(map[string]*data.AccountInfoWithStakeValues)

	for address, userID := range userAddressID {
		if delegationActiveValue, ok := userIDTotalDelegationActive[userID]; ok {
			mapAddressWithStake[address] = &data.AccountInfoWithStakeValues{
				StakeInfo: data.StakeInfo{
					DelegationLegacyActive:     delegationActiveValue.String(),
					DelegationLegacyWaitingNum: core.ComputeBalanceAsFloat(delegationActiveValue.String()),
				},
			}
		}

		if delegationWaitingValue, ok := userIDTotalDelegationWaiting[userID]; ok {
			_, found := mapAddressWithStake[address]
			if !found {
				mapAddressWithStake[address] = &data.AccountInfoWithStakeValues{
					StakeInfo: data.StakeInfo{},
				}
			}
			mapAddressWithStake[address].DelegationLegacyWaiting = delegationWaitingValue.String()
			mapAddressWithStake[address].DelegationLegacyWaitingNum = core.ComputeBalanceAsFloat(delegationWaitingValue.String())
		}

		if delegationUnStakedValue, ok := userIDTotalUnStaked[userID]; ok {
			_, found := mapAddressWithStake[address]
			if !found {
				mapAddressWithStake[address] = &data.AccountInfoWithStakeValues{
					StakeInfo: data.StakeInfo{},
				}
			}
			mapAddressWithStake[address].UnDelegateLegacy = delegationUnStakedValue.String()
			mapAddressWithStake[address].UnDelegateLegacyNum = core.ComputeBalanceAsFloat(delegationUnStakedValue.String())
		}
	}

	return mapAddressWithStake
}

func (ag *accountsGetter) extractUsersIDMap(pairsMap map[string]string) (map[string]int, error) {
	userAddressID := make(map[string]int)
	for key, value := range pairsMap {
		keyDecoded, errD := hex.DecodeString(key)
		if errD != nil {
			return nil, errD
		}

		if bytes.HasPrefix(keyDecoded, []byte(userAddressPrefix)) && len(keyDecoded) == len(userAddressPrefix)+addressLength {
			userAddress := ag.pubKeyConverter.Encode(keyDecoded[len(userAddressPrefix):])
			valueDecoded, errE := hex.DecodeString(value)
			if errE != nil {
				return nil, errE
			}
			bigValue := big.NewInt(0).SetBytes(valueDecoded)
			userAddressID[userAddress] = int(bigValue.Int64())
		}
	}

	return userAddressID, nil
}
