package process

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

const queryGetDelegatorsWithUnDelegateInfo = `{
	"query": {
		"bool": {
			"must": [
				{
					"exists": {
						"field": "unDelegateInfo"
					}
				}
			]
		}
	}
}`

type delegatorsResponse struct {
	Hits struct {
		Hits []struct {
			ID     string `json:"_id"`
			Source struct {
				Address         string `json:"address"`
				UnDelegatedInfo []struct {
					Value string `json:"value"`
				} `json:"unDelegateInfo"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type unDelegatedInfoProcessor struct {
	esClient ElasticClientHandler
}

func newUnDelegateInfoProcessor(esClient ElasticClientHandler) *unDelegatedInfoProcessor {
	return &unDelegatedInfoProcessor{
		esClient: esClient,
	}
}

func (up *unDelegatedInfoProcessor) putUnDelegateInfoFromStakingProviders(accountsWithStake map[string]*data.AccountInfoWithStakeValues) error {
	defer logExecutionTime(time.Now(), "Fetched undelegated values from staking provider contracts")
	handlerFunc := func(responseBytes []byte) error {
		delegatorsResp := &delegatorsResponse{}
		err := json.Unmarshal(responseBytes, delegatorsResp)
		if err != nil {
			return err
		}

		if len(delegatorsResp.Hits.Hits) == 0 {
			return nil
		}

		extractDataFromResponseAndPutInAccountsWithStake(delegatorsResp, accountsWithStake)

		return nil
	}

	return up.esClient.DoScrollRequestAllDocuments(dataindexer.DelegatorsIndex, []byte(queryGetDelegatorsWithUnDelegateInfo), handlerFunc)
}

func extractDataFromResponseAndPutInAccountsWithStake(delegatorsResp *delegatorsResponse, accountsWithStake map[string]*data.AccountInfoWithStakeValues) {
	for _, delegatorInfo := range delegatorsResp.Hits.Hits {
		accountWithStake, found := accountsWithStake[delegatorInfo.Source.Address]
		if !found {
			continue
		}

		undelegatedValue := big.NewInt(0)
		for _, unDelegate := range delegatorInfo.Source.UnDelegatedInfo {
			bigValue, ok := big.NewInt(0).SetString(unDelegate.Value, 10)
			if !ok {
				log.Warn("cannot convert string to big.Int", "address", accountWithStake.Address, "aunDelegate.Value", unDelegate.Value)
				continue
			}
			undelegatedValue.Add(undelegatedValue, bigValue)
		}

		setUnDelegateValue(accountWithStake, undelegatedValue)
	}
}

func setUnDelegateValue(account *data.AccountInfoWithStakeValues, undelegatedValue *big.Int) {
	if account.UnDelegateDelegation == "" {
		account.UnDelegateDelegation = undelegatedValue.String()
		account.UnDelegateDelegationNum = core.ComputeBalanceAsFloat(undelegatedValue.String())

		return
	}

	valueBig, ok := big.NewInt(0).SetString(account.UnDelegateDelegation, 10)
	if !ok {
		log.Warn("cannot convert string to big.Int", "address", account.Address, "accountWithStake.UnDelegateDelegation", account.UnDelegateDelegation)
		return
	}

	valueBig.Add(valueBig, undelegatedValue)
	account.UnDelegateDelegation = valueBig.String()
	account.UnDelegateDelegationNum = core.ComputeBalanceAsFloat(valueBig.String())
}
