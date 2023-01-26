package accountsIndexer

import (
	"bytes"
	"encoding/json"
	"fmt"

	dataIndexer "github.com/multiversx/mx-chain-es-indexer-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/tidwall/gjson"
)

const (
	numAddressesInBulk = 2000
)

var log = logger.GetOrCreate("process/accountsIndexer")

type accountsIndexer struct {
	elasticClient ElasticClientHandler
}

// NewAccountsIndexer will create a new instance of accountsIndexer
func NewAccountsIndexer(elasticClient ElasticClientHandler) (*accountsIndexer, error) {
	return &accountsIndexer{
		elasticClient: elasticClient,
	}, nil
}

// GetAccounts will get accounts by addresses from a given index
func (ai *accountsIndexer) GetAccounts(addresses []string, index string) (map[string]*data.AccountInfoWithStakeValues, error) {
	accountsES := make(map[string]*data.AccountInfoWithStakeValues)
	for idx := 0; idx < len(addresses); idx += numAddressesInBulk {
		from := idx
		to := idx + numAddressesInBulk

		var newSliceOfAddresses []string
		if to > len(addresses) {
			to = len(addresses)
			newSliceOfAddresses = make([]string, len(addresses)-idx)
		} else {
			newSliceOfAddresses = make([]string, numAddressesInBulk)
		}

		copy(newSliceOfAddresses, addresses[from:to])
		accounts, errGet := ai.getBulkOfAccounts(newSliceOfAddresses, index)
		if errGet != nil {
			log.Warn("accountsIndexer.GetAccounts: cannot get accounts", "error", errGet)
			continue
		}
		mergeAccountsMaps(accountsES, accounts)
	}

	return accountsES, nil
}

func (ai *accountsIndexer) getBulkOfAccounts(addresses []string, index string) (map[string]*data.AccountInfoWithStakeValues, error) {
	response, err := ai.elasticClient.DoMultiGet(addresses, index)
	if err != nil {
		return nil, err
	}

	accounts := make(map[string]*data.AccountInfoWithStakeValues)
	accountsResponse := make([]data.AccountsResponseES, 0)
	accountsGJson := gjson.Get(string(response), "docs")
	err = json.Unmarshal([]byte(accountsGJson.String()), &accountsResponse)
	if err != nil {
		return nil, err
	}

	for _, acct := range accountsResponse {
		if !acct.Found {
			continue
		}

		newAcct := acct.Account
		accounts[acct.ID] = &newAcct
	}

	return accounts, nil
}

// IndexAccounts will index provided accounts in a given index
func (ai *accountsIndexer) IndexAccounts(accounts map[string]*data.AccountInfoWithStakeValues, index string) error {
	buffSlice, err := serializeAccounts(accounts)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ai.elasticClient.DoBulkRequest(buffSlice[idx], index)
		if err != nil {
			return err
		}
	}

	return nil
}

func serializeAccounts(accounts map[string]*data.AccountInfoWithStakeValues) ([]*bytes.Buffer, error) {
	buffSlice := dataIndexer.NewBufferSlice(0)
	for address, acc := range accounts {
		meta, serializedData, err := prepareSerializedAccountInfo(address, acc)
		if err != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func prepareSerializedAccountInfo(
	address string,
	account *data.AccountInfoWithStakeValues,
) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, address, "\n"))
	serializedData, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}

	return meta, serializedData, nil
}

func mergeAccountsMaps(dst, src map[string]*data.AccountInfoWithStakeValues) {
	for key, value := range src {
		dst[key] = value
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ai *accountsIndexer) IsInterfaceNil() bool {
	return ai == nil
}
