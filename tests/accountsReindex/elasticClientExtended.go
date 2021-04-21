package accountsReindex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elrond-accounts-manager/elasticClient"
	"github.com/ElrondNetwork/elrond-accounts-manager/process"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/tidwall/gjson"
)

type extendedElasticClient struct {
	sourceDB      string
	destinationDB string
	process.ElasticClientHandler
}

func NewExtendedElasticClient(sourceDB, destinationDB string) (*extendedElasticClient, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{destinationDB},
		Username:  "",
		Password:  "",
	}
	dstClient, err := elasticClient.NewElasticClient(cfg)
	if err != nil {
		return nil, err
	}

	exClient := &extendedElasticClient{
		sourceDB:             sourceDB,
		destinationDB:        destinationDB,
		ElasticClientHandler: dstClient,
	}

	err = exClient.initIndex("accounts", AccountsTemplate.ToBuffer())
	if err != nil {
		return nil, nil
	}

	return exClient, nil
}

func createElasticClient(url string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
		Username:  "",
		Password:  "",
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create database reader %w", err)
	}

	return client, nil
}

func (eec *extendedElasticClient) initIndex(templateName string, template io.Reader) error {
	client, err := createElasticClient(eec.destinationDB)
	if err != nil {
		return err
	}

	res, err := client.Indices.PutTemplate(templateName, template)
	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("error response: %s", res)
	}

	defer closeBody(res)

	return nil
}

func (eec *extendedElasticClient) DoBulkRequestDestination(buff *bytes.Buffer, index string) error {
	return eec.ElasticClientHandler.DoBulkRequest(buff, index)
}

func (eec *extendedElasticClient) DoScrollRequestAllDocuments(
	index string,
	handlerFunc func(responseBytes []byte) error,
) error {
	client, err := createElasticClient(eec.sourceDB)
	if err != nil {
		return err
	}

	// use a random interval in order to avoid AWS GET request cashing
	randomNum := rand.Intn(50)
	res, err := client.Search(
		client.Search.WithSize(9000),
		client.Search.WithScroll(10*time.Minute+time.Duration(randomNum)*time.Millisecond),
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(getAll()),
	)
	if err != nil {
		return err
	}

	bodyBytes, err := getBytesFromResponse(res)
	if err != nil {
		return err
	}

	err = handlerFunc(bodyBytes)
	if err != nil {
		return err
	}

	scrollID := gjson.Get(string(bodyBytes), "_scroll_id")
	return eec.iterateScroll(client, scrollID.String(), handlerFunc)
}

func (eec *extendedElasticClient) iterateScroll(
	client *elasticsearch.Client,
	scrollID string,
	handlerFunc func(responseBytes []byte) error,
) error {
	if scrollID == "" {
		return nil
	}
	defer func() {
		err := eec.clearScroll(client, scrollID)
		if err != nil {
			log.Warn("cannot clear scroll ", err)
		}
	}()

	for {
		scrollBodyBytes, errScroll := eec.getScrollResponse(client, scrollID)
		if errScroll != nil {
			return errScroll
		}

		numberOfHits := gjson.Get(string(scrollBodyBytes), "hits.hits.#")
		if numberOfHits.Int() < 1 {
			return nil
		}
		err := handlerFunc(scrollBodyBytes)
		if err != nil {
			return err
		}
	}

}

func (eec *extendedElasticClient) getScrollResponse(client *elasticsearch.Client, scrollID string) ([]byte, error) {
	randomNum := rand.Intn(10000)
	res, err := client.Scroll(
		client.Scroll.WithScrollID(scrollID),
		client.Scroll.WithScroll(2*time.Minute+time.Duration(randomNum)*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	return getBytesFromResponse(res)
}

func getBytesFromResponse(res *esapi.Response) ([]byte, error) {
	if res.IsError() {
		return nil, fmt.Errorf("error response: %s", res)
	}
	defer closeBody(res)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func (eec *extendedElasticClient) clearScroll(client *elasticsearch.Client, scrollID string) error {
	resp, err := client.ClearScroll(
		client.ClearScroll.WithScrollID(scrollID),
	)
	if err != nil {
		return err
	}
	if resp.IsError() && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error response: %s", resp)
	}

	defer closeBody(resp)

	return nil
}

func closeBody(res *esapi.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}
