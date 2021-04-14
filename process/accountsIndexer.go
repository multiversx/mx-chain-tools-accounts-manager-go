package process

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	dataManager "github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/tidwall/gjson"
)

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
func (ai *accountsIndexer) GetAccounts(addresses []string, index string) (map[string]*data.AccountInfo, error) {
	response, err := ai.elasticClient.DoMultiGet(addresses, index)
	if err != nil {
		return nil, err
	}

	accounts := make(map[string]*data.AccountInfo)
	accountsResponse := make([]dataManager.AccountsResponseES, 0)
	accountsGJson := gjson.Get(string(response), "docs")
	err = json.Unmarshal([]byte(accountsGJson.String()), &accountsResponse)
	if err != nil {
		return nil, err
	}

	for _, acct := range accountsResponse {
		if !acct.Found {
			continue
		}

		accounts[acct.Account.Address] = &acct.Account
	}

	return accounts, nil
}

// IndexAccounts will index provided accounts in a given index
func (ai *accountsIndexer) IndexAccounts(accounts map[string]*data.AccountInfo, index string) error {
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

func serializeAccounts(accounts map[string]*data.AccountInfo) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
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
	account *data.AccountInfo,
) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, address, "\n"))
	serializedData, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}

	return meta, serializedData, nil
}
