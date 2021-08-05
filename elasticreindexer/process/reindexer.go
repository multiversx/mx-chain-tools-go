package process

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var (
	errNilElasticHandler = errors.New("nil elastic handler")
	errEmptyIndicesList  = errors.New("empty indices list to reindex")
	log                  = logger.GetOrCreate("process")
)

type reindexer struct {
	sourceElastic      ElasticClientHandler
	destinationElastic ElasticClientHandler
	indices            []string
}

// newReindexer returns a new instance of reindexer if the provided params aren't nil, or error otherwise
func newReindexer(sourceElastic ElasticClientHandler, destinationElastic ElasticClientHandler, indices []string) (*reindexer, error) {
	if check.IfNil(sourceElastic) {
		return nil, fmt.Errorf("%w for source", errNilElasticHandler)
	}
	if check.IfNil(destinationElastic) {
		return nil, fmt.Errorf("%w for destination", errNilElasticHandler)
	}
	if len(indices) == 0 {
		return nil, errEmptyIndicesList
	}

	return &reindexer{
		sourceElastic:      sourceElastic,
		destinationElastic: destinationElastic,
		indices:            indices,
	}, nil
}

// Process will handle the reindexing from source Elastic client to destination Elastic client
func (r *reindexer) Process() error {
	for _, index := range r.indices {
		err := r.processIndex(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *reindexer) processIndex(index string) error {
	originalSourceCount, err := r.sourceElastic.GetCount(index)
	if err != nil {
		return fmt.Errorf("%w while getting the source count for index %s", err, index)
	}

	err = r.copyMapping(index)
	if err != nil {
		return fmt.Errorf("%w while copying the mapping for index %s", err, index)
	}

	err = r.reindexData(index)
	if err != nil {
		return fmt.Errorf("%w while reindexing data for index %s", err, index)
	}

	destinationCount, err := r.destinationElastic.GetCount(index)
	if err != nil {
		return fmt.Errorf("%w while getting the destination count for index %s", err, index)
	}

	if destinationCount < originalSourceCount {
		return fmt.Errorf("destination elastic has %d records, while the original source count was %d", destinationCount, originalSourceCount)
	}

	return nil
}

func (r *reindexer) copyMapping(index string) error {
	sourceMapping, err := r.sourceElastic.GetMapping(index)
	if err != nil {
		return fmt.Errorf("copyMapping from source: %w", err)
	}

	err = r.destinationElastic.CreateIndexWithMapping(index, sourceMapping)
	if err != nil {
		return fmt.Errorf("copyMapping to destination: %w", err)
	}

	return nil
}

func (r *reindexer) reindexData(index string) error {
	count := 0
	handlerFunc := func(responseBytes []byte) error {
		count++
		dataBuffers, err := prepareDataForIndexing(responseBytes, index, count)
		if err != nil {
			return fmt.Errorf("%w while preparing data for indexing", err)
		}

		for i := 0; i < len(dataBuffers); i++ {
			err = r.destinationElastic.DoBulkRequest(dataBuffers[i], index)
			if err != nil {
				return fmt.Errorf("%w while r.destinationElastic.DoBulkRequest", err)
			}
		}

		return nil
	}

	err := r.sourceElastic.DoScrollRequestAllDocuments(index, getAll().Bytes(), handlerFunc)
	if err != nil {
		return fmt.Errorf("%w while r.sourceElastic.DoScrollRequestAllDocuments", err)
	}

	return nil
}

func prepareDataForIndexing(responseBytes []byte, index string, count int) ([]*bytes.Buffer, error) {
	var esResponse generalElasticResponse
	err := json.Unmarshal(responseBytes, &esResponse)
	if err != nil {
		return nil, err
	}

	resultsMap := extractSourceFromEsResponse(esResponse)
	log.Info("indexing", "index", index, "bulk size", len(resultsMap), "count", count)
	buffSlice := newBufferSlice()
	for id, source := range resultsMap {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))

		err := buffSlice.PutData(meta, source)
		if err != nil {
			return nil, err
		}

	}

	return buffSlice.Buffers(), nil
}
