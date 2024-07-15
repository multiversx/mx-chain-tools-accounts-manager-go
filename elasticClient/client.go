package elasticClient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
	"github.com/tidwall/gjson"
)

const (
	numOfErrorsToExtractBulkResponse = 5

	errPolicyAlreadyExists = "document already exists"
)

var log = logger.GetOrCreate("elasticClient")

type esClient struct {
	client      *elasticsearch.Client
	countScroll int
	clusterURL  string
}

// NewElasticClient will create a new instance of an esClient
func NewElasticClient(cfg data.EsClientConfig) (*esClient, error) {
	elasticClient, err := elasticsearch.NewClient(unWrapEsConfig(cfg))
	if err != nil {
		return nil, err
	}

	return &esClient{
		clusterURL:  cfg.Address,
		client:      elasticClient,
		countScroll: 0,
	}, nil
}

// DoBulkRequest will do a bulk of request to elastic server
func (ec *esClient) DoBulkRequest(buff *bytes.Buffer, index string) error {
	reader := bytes.NewReader(buff.Bytes())

	res, err := ec.client.Bulk(
		reader,
		ec.client.Bulk.WithIndex(index),
	)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("error DoBulkRequest: %s, url: %s", res.String(), ec.clusterURL)
	}

	defer closeBody(res)

	bodyBytes, errRead := ioutil.ReadAll(res.Body)
	if errRead != nil {
		return errRead
	}

	bulkResponse := &data.BulkRequestResponse{}
	err = json.Unmarshal(bodyBytes, bulkResponse)
	if err != nil {
		return err
	}

	if bulkResponse.Errors {
		return extractErrorFromBulkResponse(bulkResponse)
	}

	return nil
}

// DoRequest will do a index request to Elasticsearch
func (ec *esClient) DoRequest(index, documentID string, buff *bytes.Buffer) error {
	req := &esapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       buff,
	}

	res, err := req.Do(context.Background(), ec.client)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("error DoRequest: %s", res.String())
	}

	defer closeBody(res)

	return nil
}

// CheckIfIndexExists will check if an index exists
func (ec *esClient) CheckIfIndexExists(index string) (bool, error) {
	res, err := ec.client.Indices.Exists(
		[]string{index},
	)
	if err != nil {
		return false, err
	}

	return exists(res), nil
}

func exists(res *esapi.Response) bool {
	defer func() {
		if res != nil && res.Body != nil {
			err := res.Body.Close()
			if err != nil {
				log.Warn("elasticClient.exists", "could not close body: ", err.Error())
			}
		}
	}()

	switch res.StatusCode {
	case http.StatusOK:
		return true
	case http.StatusNotFound:
		return false
	default:
		log.Warn("elasticClient.exists", "invalid status code returned by the elastic nodes:", res.StatusCode)
		return false
	}
}

// DoMultiGet wil do a multi get request to elaticsearch server
func (ec *esClient) DoMultiGet(ids []string, index string) ([]byte, error) {
	buff := getDocumentsByIDsQueryEncoded(ids)
	res, err := ec.client.Mget(
		buff,
		ec.client.Mget.WithIndex(index),
	)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("error DoMultiGet: %s", res.String())
	}

	defer closeBody(res)

	bodyBytes, errRead := ioutil.ReadAll(res.Body)
	if errRead != nil {
		return nil, errRead
	}

	return bodyBytes, nil
}

// PutMapping will put mapping for a given index
func (ec *esClient) PutMapping(targetIndex string, body *bytes.Buffer) error {
	res, err := ec.client.Indices.PutMapping(
		body,
		ec.client.Indices.PutMapping.WithIndex(targetIndex),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("error PutMapping: %s", res.String())
	}

	defer closeBody(res)

	return nil
}

// CreateIndexWithMapping will init an index and put the template
func (ec *esClient) CreateIndexWithMapping(index string, mapping *bytes.Buffer) error {
	res, err := ec.client.Indices.Create(
		index,
		ec.client.Indices.Create.WithBody(mapping),
	)

	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("error CreateIndexWithMapping: %s, url: %s", res.String(), ec.clusterURL)
	}

	defer closeBody(res)

	return nil
}

// PutPolicy will put in Elasticsearch cluster the provided policy with the given name
func (ec *esClient) PutPolicy(policyName string, policy *bytes.Buffer) error {
	res, err := ec.client.ILM.PutLifecycle(
		policyName,
		ec.client.ILM.PutLifecycle.WithBody(policy),
	)
	if err != nil {
		return err
	}

	bodyBytes, errGet := getBytesFromResponse(res)
	if errGet != nil {
		return errGet
	}

	response := &responseCreatePolicy{}
	err = json.Unmarshal(bodyBytes, response)
	if err != nil {
		return err
	}

	errStr := fmt.Sprintf("%v", response.Error)
	if response.Status == http.StatusConflict && !strings.Contains(errStr, errPolicyAlreadyExists) {
		return fmt.Errorf("error esClient.PutPolicy: %s", errStr)
	}

	return nil
}

// DoScrollRequestAllDocuments will perform a documents request using scroll api
func (ec *esClient) DoScrollRequestAllDocuments(
	index string,
	body []byte,
	handlerFunc func(responseBytes []byte) error,
) error {
	ec.countScroll++
	res, err := ec.client.Search(
		ec.client.Search.WithSize(9000),
		ec.client.Search.WithScroll(2*time.Hour+time.Duration(ec.countScroll)*time.Millisecond),
		ec.client.Search.WithContext(context.Background()),
		ec.client.Search.WithIndex(index),
		ec.client.Search.WithBody(bytes.NewBuffer(body)),
	)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("error DoScrollRequestAllDocuments: %s", res.String())
	}

	bodyBytes, errGet := getBytesFromResponse(res)
	if errGet != nil {
		return errGet
	}

	err = handlerFunc(bodyBytes)
	if err != nil {
		return err
	}

	scrollID := gjson.Get(string(bodyBytes), "_scroll_id")
	return ec.iterateScroll(scrollID.String(), handlerFunc)
}

func (ec *esClient) iterateScroll(
	scrollID string,
	handlerFunc func(responseBytes []byte) error,
) error {
	if scrollID == "" {
		return nil
	}
	defer func() {
		err := ec.clearScroll(scrollID)
		if err != nil {
			log.Warn("cannot clear scroll", "error", err)
		}
	}()

	for {
		scrollBodyBytes, errScroll := ec.getScrollResponse(scrollID)
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

func (ec *esClient) getScrollResponse(scrollID string) ([]byte, error) {
	ec.countScroll++
	res, err := ec.client.Scroll(
		ec.client.Scroll.WithScrollID(scrollID),
		ec.client.Scroll.WithScroll(2*time.Minute+time.Duration(ec.countScroll)*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	return getBytesFromResponse(res)
}

func (ec *esClient) clearScroll(scrollID string) error {
	resp, err := ec.client.ClearScroll(
		ec.client.ClearScroll.WithScrollID(scrollID),
	)
	if err != nil {
		return err
	}
	if resp.IsError() && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error response: %s", resp.String())
	}

	defer closeBody(resp)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ec *esClient) IsInterfaceNil() bool {
	return ec == nil
}
