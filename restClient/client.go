package restClient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/core"
	"github.com/multiversx/mx-chain-tools-accounts-manager-go/data"
)

const maxNumOfRetries = 10

var log = logger.GetOrCreate("restClient")

type restClient struct {
	httpClient *http.Client
	url        string
}

// NewRestClient will create a new instance of restClient
func NewRestClient(url string) (*restClient, error) {
	c := http.DefaultClient

	return &restClient{
		httpClient: c,
		url:        url,
	}, nil
}

// CallGetRestEndPoint calls an external end point (sends a get request)
func (rc *restClient) CallGetRestEndPoint(
	path string,
	value interface{},
	authenticationData data.RestApiAuthenticationData,
) error {
	req, err := http.NewRequest("GET", rc.url+path, nil)
	if err != nil {
		return err
	}

	userAgent := "Accounts manager>"
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if core.ShouldUseBasicAuthentication(authenticationData) {
		req.SetBasicAuth(authenticationData.Username, authenticationData.Password)
	}

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		errNotCritical := resp.Body.Close()
		if errNotCritical != nil {
			log.Warn("restClient.CallGetRestEndPoint: close body", "error", errNotCritical.Error())
		}
	}()

	err = json.NewDecoder(resp.Body).Decode(value)
	if err != nil {
		return err
	}

	return nil
}

// CallPostRestEndPoint calls an external end point (sends a post request)
func (rc *restClient) CallPostRestEndPoint(
	path string,
	dataR interface{},
	response interface{},
	authenticationData data.RestApiAuthenticationData,
) error {
	buff, err := json.Marshal(dataR)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", rc.url+path, bytes.NewReader(buff))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "userAgent")
	if core.ShouldUseBasicAuthentication(authenticationData) {
		req.SetBasicAuth(authenticationData.Username, authenticationData.Password)
	}

	var count int
	var resp *http.Response
	for {
		resp, err = rc.httpClient.Do(req)

		if err != nil {
			if count < maxNumOfRetries {
				log.Warn("rc.httpClient.Do", "error", err)
				count++
				sleep(count)
				continue
			}
			return fmt.Errorf("too many retries, error: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusRequestTimeout {
			_ = resp.Body.Close()
			if count < maxNumOfRetries {
				count++
				sleep(count)
				continue
			}
		}

		break
	}

	defer func() {
		errNotCritical := resp.Body.Close()
		if errNotCritical != nil {
			log.Warn("restClient.CallPostRestEndPoint: close body", "error", errNotCritical.Error())
		}
	}()

	responseStatusCode := resp.StatusCode
	if responseStatusCode == http.StatusOK { // everything ok, return status ok and the expected response
		return json.NewDecoder(resp.Body).Decode(response)
	}

	// status response not ok, return the error
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	genericApiResponse := data.GenericAPIResponse{}
	err = json.Unmarshal(responseBytes, &genericApiResponse)
	if err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return errors.New(genericApiResponse.Error)
}

func sleep(count int) {
	delay := time.Duration(math.Exp2(float64(count))) * time.Second
	time.Sleep(delay)
}
