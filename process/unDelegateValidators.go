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

const (
	maxNumberOfParallelRequests   = 40
	getUnStakedTokensListEndpoint = "getUnStakedTokensList"
)

func (ag *accountsGetter) putUndelegatedValuesFromValidatorsContract(accountsWithStake map[string]*data.AccountInfoWithStakeValues) error {
	if ag.validatorsContract == "" {
		return nil
	}

	unDelegatedValues, err := ag.getUnDelegatedValuesFromValidatorsContract(accountsWithStake)
	if err != nil {
		return err
	}

	for address, unDelegatedValue := range unDelegatedValues {
		if _, found := accountsWithStake[address]; !found {
			continue
		}

		accountsWithStake[address].UnDelegateValidator = unDelegatedValue.String()
		accountsWithStake[address].UnDelegateValidatorNum = core.ComputeBalanceAsFloat(unDelegatedValue.String())
	}

	return nil
}

func (ag *accountsGetter) getUnDelegatedValuesFromValidatorsContract(accountsWithStake map[string]*data.AccountInfoWithStakeValues) (map[string]*big.Int, error) {
	defer logExecutionTime(time.Now(), "Fetched undelegated values from validators contract")

	unDelegatedValue := make(map[string]*big.Int)

	done, wg := make(chan struct{}, maxNumberOfParallelRequests), &sync.WaitGroup{}
	errors := make([]string, 0)
	for address := range accountsWithStake {
		done <- struct{}{}
		wg.Add(1)

		go func(addr string) {
			defer func() {
				<-done
				wg.Done()
			}()

			value, err := ag.getUnDelegatedValueForAddressValidatorsContract(addr)
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

func (ag *accountsGetter) getUnDelegatedValueForAddressValidatorsContract(address string) (*big.Int, error) {

	decodedAddr, err := ag.pubKeyConverter.Decode(address)
	if err != nil {
		return nil, err
	}

	vmRequest := &data.VmValueRequest{
		Address:    ag.validatorsContract,
		FuncName:   getUnStakedTokensListEndpoint,
		CallerAddr: ag.validatorsContract,
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

	if responseVmValue.Data.Data == nil {
		return big.NewInt(0), nil
	}

	returnData := responseVmValue.Data.Data.ReturnData
	if len(returnData) == 0 {
		return big.NewInt(0), nil
	}

	step := 2
	value := big.NewInt(0)
	for idx := 0; idx < len(returnData); idx += step {
		value.Add(value, big.NewInt(0).SetBytes(returnData[idx]))
	}

	return value, nil
}
