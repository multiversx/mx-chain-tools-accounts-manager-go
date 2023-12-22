package process

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const getUserStakeByTypeEndpoint = "getUserStakeByType"

func (ag *accountsGetter) putUnDelegatedValuesFromDelegationLegacy(accountsWithStake map[string]*data.AccountInfoWithStakeValues) error {
	unDelegatedValues, err := ag.getUnDelegatedValuesFromDelegationLegacyContract(accountsWithStake)
	if err != nil {
		return err
	}

	for address, unDelegatedValue := range unDelegatedValues {
		if _, found := accountsWithStake[address]; !found {
			continue
		}

		accountsWithStake[address].UnDelegateLegacy = unDelegatedValue.String()
		accountsWithStake[address].UnDelegateLegacyNum = core.ComputeBalanceAsFloat(unDelegatedValue.String())
	}

	return nil
}

func (ag *accountsGetter) getUnDelegatedValuesFromDelegationLegacyContract(accountsWithStake map[string]*data.AccountInfoWithStakeValues) (map[string]*big.Int, error) {
	defer logExecutionTime(time.Now(), "Fetched undelegated values from delegation legacy contract")

	unDelegatedValue := make(map[string]*big.Int)

	done, wg := make(chan struct{}, maxNumberOfParallelRequests), &sync.WaitGroup{}
	errors := make([]string, 0)
	for address := range accountsWithStake {
		done <- struct{}{}
		wg.Add(1)

		go func(addr string) {
			value, err := ag.getUnDelegatedValueForAddressDelegationLegacyContract(addr, done, wg)
			if err != nil {
				ag.mutex.Lock()
				errors = append(errors, err.Error())
				ag.mutex.Unlock()
				return
			}
			if big.NewInt(0).Cmp(value) == 0 {
				return
			}

			ag.mutex.Lock()
			unDelegatedValue[addr] = value
			ag.mutex.Unlock()
		}(address)

	}

	wg.Wait()
	if len(errors) > 0 {
		return nil, fmt.Errorf("%s", errors[0])
	}

	return unDelegatedValue, nil
}

func (ag *accountsGetter) getUnDelegatedValueForAddressDelegationLegacyContract(address string, done chan struct{}, wg *sync.WaitGroup) (*big.Int, error) {
	defer func() {
		<-done
		wg.Done()
	}()

	decodedAddr, err := ag.pubKeyConverter.Decode(address)
	if err != nil {
		return nil, err
	}

	vmRequest := &data.VmValueRequest{
		Address:    ag.delegationContractAddress,
		FuncName:   getUserStakeByTypeEndpoint,
		CallerAddr: ag.delegationContractAddress,
		Args:       []string{hex.EncodeToString(decodedAddr)},
	}

	responseVmValue := &data.ResponseVmValue{}
	err = ag.restClient.CallPostRestEndPoint(pathVMValues, vmRequest, responseVmValue, ag.authenticationData)
	if err != nil {
		return nil, err
	}
	if responseVmValue.Error != "" {
		return nil, fmt.Errorf("%s", responseVmValue.Error)
	}
	if responseVmValue.Data.Data != nil {
		if responseVmValue.Data.Data.ReturnCode != vmcommon.Ok.String() {
			return nil, fmt.Errorf("%s: %s", responseVmValue.Data.Data.ReturnCode, responseVmValue.Data.Data.ReturnMessage)
		}
	}

	returnData := responseVmValue.Data.Data.ReturnData
	if len(returnData) == 0 {
		return big.NewInt(0), nil
	}

	if len(returnData) < 5 {
		return big.NewInt(0), nil
	}

	value := big.NewInt(0).SetBytes(returnData[3])
	value.Add(value, big.NewInt(0).SetBytes(returnData[4]))

	if core.ComputeBalanceAsFloat(value.String()) < 0.001 {
		return big.NewInt(0), nil
	}

	return value, nil
}
