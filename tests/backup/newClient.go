package accountsReindex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/ElrondNetwork/elrond-accounts-manager/elasticClient"
	"github.com/ElrondNetwork/elrond-accounts-manager/process"
	"github.com/elastic/go-elasticsearch/v7"
)

type object = map[string]interface{}

type esClient struct {
	client *elasticsearch.Client
	ec     process.ElasticClientHandler
}

// NewElasticSearchClient will create a new instance of elastic search client
func NewElasticSearchClient(from string, to string) (*esClient, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{from},
		Username:  "",
		Password:  "",
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create database reader %w", err)
	}

	elasticCfg := elasticsearch.Config{
		Addresses: []string{to},
	}
	ec, err := elasticClient.NewElasticClient(elasticCfg)
	if err != nil {
		return nil, err
	}

	elasticC := &esClient{
		client: client,
		ec:     ec,
	}

	err = createIndexTemplate(to, "accounts", AccountsTemplate.ToBuffer())
	if err != nil {
		return nil, err
	}

	return elasticC, nil
}

// CreateIndexTemplate creates an elasticsearch index template
func createIndexTemplate(to string, templateName string, template io.Reader) error {
	elasticCfg := elasticsearch.Config{
		Addresses: []string{to},
	}
	eeee, err := elasticsearch.NewClient(elasticCfg)
	if err != nil {
		return err
	}

	res, err := eeee.Indices.PutTemplate(templateName, template)
	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("error response: %s", res)
	}

	return nil
}

func encodeQuery(query object) (bytes.Buffer, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(query); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error encoding query: %s", err.Error())
	}

	return buff, nil
}

func myQuery() object {
	return object{
		"query": object{
			"match_all": object{},
		},
	}
}

func (ec *esClient) ProcessAllAccounts() error {
	buff, err := encodeQuery(myQuery())
	if err != nil {
		return err
	}

	// use a random interval in order to avoid AWS GET request cashing
	randomNum := rand.Intn(50)
	res, err := ec.client.Search(
		ec.client.Search.WithSize(9000),
		ec.client.Search.WithScroll(10*time.Minute+time.Duration(randomNum)*time.Millisecond),
		ec.client.Search.WithContext(context.Background()),
		ec.client.Search.WithIndex("accounts"),
		ec.client.Search.WithBody(&buff),
	)

	defer func() {
		if res != nil {
			_ = res.Body.Close()
		}
	}()

	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("error response: %s", res)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	responseScroll := &AllAccountsResponse{}
	err = json.Unmarshal(bodyBytes, &responseScroll)
	if err != nil {
		return err
	}

	allAccounts := make(map[string]*data.AccountInfoWithStakeValues)
	getAllAccounts(allAccounts, responseScroll)

	err = ec.iterateScroll(responseScroll.ScrollID, allAccounts)
	if err != nil {
		return err
	}

	return nil
}

var count = 0

func getAllAccounts(mallTxs map[string]*data.AccountInfoWithStakeValues, response *AllAccountsResponse) {
	for _, acct := range response.Hits.Hits {

		acc := data.AccountInfoWithStakeValues{}

		acc = acct.Account

		if acc.BalanceNum > 0 {
			count++
		}

		mallTxs[acct.ID] = &acc
	}
}

func (ec *esClient) iterateScroll(scrollID string, mapAllAccounts map[string]*data.AccountInfoWithStakeValues) error {
	defer func() {
		err := ec.clearScroll(scrollID)
		if err != nil {
			log.Fatal("cannot clear scroll ", err)
		}
	}()

	for {
		scrollBodyBytes, errScroll := ec.getScrollResponse(scrollID)
		if errScroll != nil {
			return errScroll
		}

		responseScroll := &AllAccountsResponse{}
		err := json.Unmarshal(scrollBodyBytes, &responseScroll)
		if err != nil {
			return err
		}

		if len(responseScroll.Hits.Hits) < 1 {
			fmt.Println("Total account   ", count)
			break
		}

		getAllAccounts(mapAllAccounts, responseScroll)

		err = ec.indexAllAccounts(mapAllAccounts)
		if err != nil {
			return err
		}

		mapAllAccounts = make(map[string]*data.AccountInfoWithStakeValues)

	}

	return nil
}

func (ec *esClient) indexAllAccounts(mapAllAccounts map[string]*data.AccountInfoWithStakeValues) error {
	acIndexer, err := process.NewAccountsIndexer(ec.ec)
	if err != nil {
		return err
	}

	return acIndexer.IndexAccounts(mapAllAccounts, "accounts-000001")
}

func (ec *esClient) getScrollResponse(scrollID string) ([]byte, error) {
	res, err := ec.client.Scroll(
		ec.client.Scroll.WithScrollID(scrollID),
		ec.client.Scroll.WithScroll(2*time.Minute),
	)

	defer func() {
		if res != nil {
			_ = res.Body.Close()
		}
	}()

	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("error response: %s", res)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func (ec *esClient) clearScroll(scrollID string) error {
	resp, err := ec.client.ClearScroll(
		ec.client.ClearScroll.WithScrollID(scrollID),
	)

	if err != nil {
		return err
	}

	if resp.IsError() && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error response: %s", resp)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	return nil
}
