package elasticClient

import (
	"fmt"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-accounts-manager/data"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// Response is a structure that holds response from Kibana
type responseCreatePolicy struct {
	Error  interface{} `json:"error,omitempty"`
	Status int         `json:"status"`
}

func unWrapEsConfig(wrappedConfig data.EsClientConfig) elasticsearch.Config {
	return elasticsearch.Config{
		Addresses: []string{wrappedConfig.Address},
		Username:  wrappedConfig.Username,
		Password:  wrappedConfig.Password,
	}
}

func closeBody(res *esapi.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}

func getBytesFromResponse(res *esapi.Response) ([]byte, error) {
	if res.IsError() {
		return nil, fmt.Errorf("error response: %s", res.String())
	}
	defer closeBody(res)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func extractErrorFromBulkResponse(response *data.BulkRequestResponse) error {
	count := 0
	errorsString := ""
	for _, item := range response.Items {
		if item.Index.Status < 300 {
			continue
		}

		count++
		errorsString += fmt.Sprintf("{ status code: %d, error type: %s, reason: %s }\n", item.Index.Status, item.Index.Error.Type, item.Index.Error.Reason)

		if count == numOfErrorsToExtractBulkResponse {
			break
		}
	}

	log.Warn("extractErrorFromBulkResponse", "error", errorsString)
	return fmt.Errorf("%s", errorsString)
}
