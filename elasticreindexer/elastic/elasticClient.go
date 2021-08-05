package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/tidwall/gjson"
)

var log = logger.GetOrCreate("elastic")

const stepDelayBetweenRequests = 1 * time.Second

type esClient struct {
	client *elasticsearch.Client

	// countScroll is used to be incremented after each scroll so the scroll duration is different each time,
	// bypassing any possible caching based on the same request
	countScroll int
}

// NewElasticClient will create a new instance of an esClient
func NewElasticClient(cfg config.ElasticInstanceConfig) (*esClient, error) {
	elasticClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.URL},
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		return nil, err
	}

	return &esClient{
		client:      elasticClient,
		countScroll: 0,
	}, nil
}

// GetCount returns the total number of documents available in the provided index
func (esc *esClient) GetCount(index string) (uint64, error) {
	res, err := esc.client.Count(
		esc.client.Count.WithIndex(index),
	)
	if err != nil {
		return 0, err
	}

	respBytes, err := getBytesFromResponse(res)
	if err != nil {
		return 0, err
	}

	count := gjson.Get(string(respBytes), "count")
	return count.Uint(), nil
}

// GetMapping will return the mapping of the specified index
func (esc *esClient) GetMapping(index string) (*bytes.Buffer, error) {
	res, err := esc.client.Indices.GetMapping(
		esc.client.Indices.GetMapping.WithIndex(index),
	)
	if err != nil {
		return nil, err
	}

	respBytes, err := getBytesFromResponse(res)
	if err != nil {
		return nil, err
	}

	propertiesRes := gjson.Get(string(respBytes), fmt.Sprintf("%s-000001", index))

	return bytes.NewBufferString(propertiesRes.Raw), nil
}

// CreateIndexWithMapping will create an index with the provided
func (esc *esClient) CreateIndexWithMapping(targetIndex string, body *bytes.Buffer) error {
	res, err := esc.client.Indices.Create(
		targetIndex,
		esc.client.Indices.Create.WithBody(body),
	)
	if err != nil {
		return err
	}

	defer closeBody(res)

	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return nil
}

// DoScrollRequestAllDocuments will perform a documents request using scroll api
func (esc *esClient) DoScrollRequestAllDocuments(
	index string,
	body []byte,
	handlerFunc func(responseBytes []byte) error,
) error {
	esc.countScroll++
	res, err := esc.client.Search(
		esc.client.Search.WithSize(9000),
		esc.client.Search.WithScroll(10*time.Minute+time.Duration(esc.countScroll)*time.Millisecond),
		esc.client.Search.WithContext(context.Background()),
		esc.client.Search.WithIndex(index),
		esc.client.Search.WithBody(bytes.NewBuffer(body)),
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
	return esc.iterateScroll(scrollID.String(), handlerFunc)
}

// DoBulkRequest will do a bulk of request to elastic server
func (esc *esClient) DoBulkRequest(buff *bytes.Buffer, index string) error {
	reader := bytes.NewReader(buff.Bytes())

	res, err := esc.client.Bulk(
		reader,
		esc.client.Bulk.WithIndex(index),
	)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	defer closeBody(res)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	bulkResponse := &bulkRequestResponse{}
	err = json.Unmarshal(bodyBytes, bulkResponse)
	if err != nil {
		return err
	}

	if bulkResponse.Errors {
		return extractErrorFromBulkResponse(bulkResponse)
	}

	return nil
}

func (esc *esClient) iterateScroll(
	scrollID string,
	handlerFunc func(responseBytes []byte) error,
) error {
	if scrollID == "" {
		return nil
	}
	defer func() {
		err := esc.clearScroll(scrollID)
		if err != nil {
			log.Warn("cannot clear scroll", "error", err)
		}
	}()

	for {
		scrollBodyBytes, errScroll := esc.getScrollResponse(scrollID)
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

		time.Sleep(stepDelayBetweenRequests)
	}
}

func (esc *esClient) getScrollResponse(scrollID string) ([]byte, error) {
	esc.countScroll++
	res, err := esc.client.Scroll(
		esc.client.Scroll.WithScrollID(scrollID),
		esc.client.Scroll.WithScroll(2*time.Minute+time.Duration(esc.countScroll)*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	return getBytesFromResponse(res)
}

func (esc *esClient) clearScroll(scrollID string) error {
	resp, err := esc.client.ClearScroll(
		esc.client.ClearScroll.WithScrollID(scrollID),
	)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if resp.IsError() && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error response: %s", resp)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (esc *esClient) IsInterfaceNil() bool {
	return esc == nil
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

func closeBody(res *esapi.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}
