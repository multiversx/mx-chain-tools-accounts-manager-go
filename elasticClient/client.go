package elasticClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/elastic/go-elasticsearch/v7"
)

const numOfErrorsToExtractBulkResponse = 5

type esClient struct {
	client *elasticsearch.Client
}

// NewElasticClient will create a new instance of an esClient
func NewElasticClient(cfg elasticsearch.Config) (*esClient, error) {
	elasticClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &esClient{
		client: elasticClient,
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
		return fmt.Errorf("%s", res.String())
	}

	defer func() {
		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}
	}()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
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

func extractErrorFromBulkResponse(response *data.BulkRequestResponse) error {
	count := 0
	errorsString := ""
	for _, item := range response.Items {
		count++
		errorsString += fmt.Sprintf("{ status code: %d, error type: %s, reason: %s }\n", item.Index.Status, item.Index.Error.Type, item.Index.Error.Reason)

		if count == numOfErrorsToExtractBulkResponse {
			break
		}
	}

	return fmt.Errorf("%s", errorsString)
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
		return nil, fmt.Errorf("%s", res.String())
	}

	defer func() {
		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}
	}()

	bodyBytes, errRead := ioutil.ReadAll(res.Body)
	if errRead != nil {
		return nil, errRead
	}

	return bodyBytes, nil
}

// WaitYellowStatus will wait for yellow status of the ES cluster (wait clone operation to be done)
func (ec *esClient) WaitYellowStatus() error {
	res, err := ec.client.Cluster.Health(
		ec.client.Cluster.Health.WithWaitForStatus("yellow"),
	)

	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return nil
}

// CloneIndex wil try to clone an index
// to clone an index we have to set first index as "read-only" and after that do clone operation
// after the clone operation is done we have to used read-only setting
// this function return if index was cloned and an error
func (ec *esClient) CloneIndex(index, targetIndex string) (cloned bool, err error) {
	err = ec.setReadOnly(index)
	if err != nil {
		return
	}

	defer func() {
		errUnset := ec.UnsetReadOnly(index)
		if err != nil && errUnset != nil {
			err = fmt.Errorf("error clone: %w, error unsetReadOnly: %s", err, errUnset)
			return
		}
		return
	}()

	res, errClone := ec.client.Indices.Clone(
		index,
		targetIndex,
		ec.client.Indices.Clone.WithWaitForActiveShards("1"),
	)
	if errClone != nil {
		err = errClone
		return
	}

	if res.IsError() {
		err = fmt.Errorf("%s", res.String())
		return
	}

	cloned = true
	return
}

// UnsetReadOnly will unset property "read-only" of an elasticsearch index
func (ec *esClient) UnsetReadOnly(index string) error {
	return ec.putSettings(false, index)
}

func (ec *esClient) setReadOnly(index string) error {
	return ec.putSettings(true, index)
}

func (ec *esClient) putSettings(readOnly bool, index string) error {
	buff := settingsWriteEncoded(readOnly)

	res, err := ec.client.Indices.PutSettings(
		buff,
		ec.client.Indices.PutSettings.WithIndex(index),
	)
	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return nil
}
