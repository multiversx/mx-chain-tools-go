package process

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

const (
	scDeploysIndex    = "scdeploys"
	transactionsIndex = "transactions"
)

type scCounter struct {
	esClient ElasticClientHandler

	wasmBytes       []byte
	contractDeploys uint64
	users           map[string]uint64
}

var (
	errNilElasticHandler = errors.New("nil elastic handler")
)

func newSCCounter(esClient ElasticClientHandler) (*scCounter, error) {
	if check.IfNil(esClient) {
		return nil, errNilElasticHandler
	}

	return &scCounter{
		esClient: esClient,
		users:    map[string]uint64{},
	}, nil
}

func (scc *scCounter) ProcessSCWasm(pathToWasmFile string) error {
	scCodeHex, err := getSCCode(pathToWasmFile)
	if err != nil {
		return err
	}

	scc.wasmBytes = []byte(scCodeHex)

	err = scc.esClient.DoScrollRequestAllDocuments(scDeploysIndex, getAll().Bytes(), scc.scDeploysHandler)
	if err != nil {
		return err
	}

	fmt.Printf("\nNumber of contract deploys: %d\n", scc.contractDeploys)

	scc.displayDeployers()

	return nil
}

func (scc *scCounter) displayDeployers() {
	for user, numDeploys := range scc.users {
		if numDeploys == 0 {
			continue
		}

		fmt.Printf("user:%s num-deploys:%d\n", user, numDeploys)
	}
}

func (scc *scCounter) scDeploysHandler(response []byte) error {
	scDeploysResponse := &scDeploysElasticResponse{}
	err := json.Unmarshal(response, scDeploysResponse)
	if err != nil {
		return err
	}

	deploysHashes := make([]string, 0, len(scDeploysResponse.Hits.Hits))
	for _, hit := range scDeploysResponse.Hits.Hits {
		if hit.Source.TxHash == "" {
			continue
		}
		deploysHashes = append(deploysHashes, hit.Source.TxHash)
		scc.users[hit.Source.Creator] = 0
	}

	query := getDocumentsByIDsQuery(deploysHashes).Bytes()
	return scc.esClient.DoScrollRequestAllDocuments(transactionsIndex, query, scc.transactionsHandler)
}

func (scc *scCounter) transactionsHandler(response []byte) error {
	transactionsResponse := &transactionsElasticResponse{}
	err := json.Unmarshal(response, transactionsResponse)
	if err != nil {
		return err
	}

	for _, hit := range transactionsResponse.Hits.Hits {
		dataField := hit.Source.Data
		if bytes.Contains(dataField, scc.wasmBytes) {
			scc.contractDeploys++
			scc.users[hit.Source.Sender]++
		}
	}

	return nil
}

func getSCCode(fileName string) (string, error) {
	code, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return "", fmt.Errorf("cannot get smart contract code %w", err)
	}

	encodeEncoded := hex.EncodeToString(code)
	return encodeEncoded, nil
}
