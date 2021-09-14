package process

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

type object = map[string]interface{}

func encodeQuery(query object) (bytes.Buffer, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(query); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error encoding query: %s", err.Error())
	}

	return buff, nil
}

func getAll() *bytes.Buffer {
	obj := object{
		"query": object{
			"match_all": object{},
		},
	}

	encoded, _ := encodeQuery(obj)

	return &encoded
}

type scDeploysElasticResponse struct {
	Hits struct {
		Hits []struct {
			ID     string            `json:"_id"`
			Source data.ScDeployInfo `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type transactionsElasticResponse struct {
	Hits struct {
		Hits []struct {
			ID     string           `json:"_id"`
			Source data.Transaction `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func getDocumentsByIDsQuery(hashes []string) *bytes.Buffer {
	obj := object{
		"query": object{
			"terms": object{
				"_id": hashes,
			},
		},
	}

	encoded, _ := encodeQuery(obj)

	return &encoded
}
