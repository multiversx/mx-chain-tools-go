package process

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var (
	errNilElasticHandler = errors.New("nil elastic handler")
	log                  = logger.GetOrCreate("process")
)

const indexSuffix = "-000001"

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

	return &reindexer{
		sourceElastic:      sourceElastic,
		destinationElastic: destinationElastic,
		indices:            indices,
	}, nil
}

// Process will handle the reindexing from source Elastic client to destination Elastic client
func (r *reindexer) Process(overwrite bool, skipMappings bool, index string, newName string) error {
	return r.processIndex(index, overwrite, skipMappings, newName)
}

func (r *reindexer) processIndex(index string, overwrite bool, skipMappings bool, newName string) error {
	originalSourceCount, err := r.sourceElastic.GetCount(index)
	if err != nil {
		return fmt.Errorf("%w while getting the source count for index %s", err, index)
	}

	err = r.copyMappingIfNecessary(index, overwrite, skipMappings, newName)
	if err != nil {
		return fmt.Errorf("%w while copying the mapping for index %s", err, index)
	}

	log.Info("starting reindexing", "index", index, "index new name", newName)

	err = r.reindexData(index, newName)
	if err != nil {
		return fmt.Errorf("%w while reindexing data for index %s", err, index)
	}

	destinationCount, err := r.destinationElastic.GetCount(index)
	if err != nil {
		return fmt.Errorf("%w while getting the destination count for index %s", err, index)
	}

	log.Info("finished indexing for index",
		"index", index,
		"original source count", originalSourceCount,
		"destination count (estimation)", destinationCount)

	return nil
}

func (r *reindexer) copyMappingIfNecessary(index string, overwrite bool, skipMappings bool, newIndexName string) error {
	if skipMappings {
		return nil
	}

	indexxxxx := index
	indexWithSuffix := index + indexSuffix
	newIndexNameWithSuffix := newIndexName
	somethingNew := indexWithSuffix
	if newIndexName != "" {
		newIndexNameWithSuffix += indexSuffix
		somethingNew = newIndexNameWithSuffix
		indexxxxx = newIndexName
	}

	aliasExists := r.destinationElastic.DoesAliasExist(indexxxxx)
	if aliasExists && !overwrite {
		return fmt.Errorf("index with alias %s already exists. Please clean the destination indexer before"+
			" retrying, or start the tool using --overwrite flag", index)
	}

	indexExists := r.destinationElastic.DoesIndexExist(somethingNew)
	if indexExists && !overwrite {
		return fmt.Errorf("index %s already exists. Please clean the destination indexer before"+
			" retrying, or start the tool using --overwrite flag", index)
	}

	if !indexExists {
		sourceMapping, err := r.sourceElastic.GetMapping(index)
		if err != nil {
			return fmt.Errorf("error while getting mapping from source: %w", err)
		}

		sourceMappingStr := sourceMapping.String()
		if newIndexName != "" {
			sourceMappingStr = strings.Replace(sourceMappingStr, indexWithSuffix, newIndexNameWithSuffix, 1)
		}

		err = r.destinationElastic.CreateIndexWithMapping(somethingNew, bytes.NewBuffer([]byte(sourceMappingStr)))
		if err != nil {
			return fmt.Errorf("error while creating index with mapping to destination: %w", err)
		}
	}

	if aliasExists {
		return nil
	}

	return r.destinationElastic.PutAlias(somethingNew, indexxxxx)
}

func (r *reindexer) reindexData(index string, newIndex string) error {
	if newIndex == "" {
		newIndex = index
	}

	count := 0
	handlerFunc := func(responseBytes []byte) error {
		count++
		dataBuffers, err := prepareDataForIndexing(responseBytes, newIndex, count)
		if err != nil {
			return fmt.Errorf("%w while preparing data for indexing", err)
		}

		for i := 0; i < len(dataBuffers); i++ {
			err = r.destinationElastic.DoBulkRequest(dataBuffers[i], newIndex)
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
	log.Info("\tindexing", "index", index, "bulk size", len(resultsMap), "count", count)
	buffSlice := newBufferSlice()
	for id, source := range resultsMap {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))

		err = buffSlice.PutData(meta, source)
		if err != nil {
			return nil, err
		}

	}

	return buffSlice.Buffers(), nil
}

// ProcessIndexWithTimestamp will handle the reindexing from source Elastic client to destination Elastic client based on the provided interval
func (r *reindexer) ProcessIndexWithTimestamp(index string, overwrite bool, skipMappings bool, start, stop int64, count *uint64) error {
	err := r.copyMappingIfNecessary(index, overwrite, skipMappings, "")
	if err != nil {
		return fmt.Errorf("%w while copying the mapping for index %s", err, index)
	}

	scrollRequestHandlerFunc := r.createScrollRequestHandlerFunction(count, index)
	err = r.sourceElastic.DoScrollRequestAllDocuments(index, getWithTimestamp(start, stop).Bytes(), scrollRequestHandlerFunc)
	if err != nil {
		return fmt.Errorf("%w while r.sourceElastic.DoScrollRequestAllDocuments", err)
	}

	return nil
}

func (r *reindexer) createScrollRequestHandlerFunction(count *uint64, index string) func([]byte) error {
	return func(responseBytes []byte) error {
		atomic.AddUint64(count, 1)
		dataBuffers, errP := prepareDataForIndexing(responseBytes, index, int(atomic.LoadUint64(count)))
		if errP != nil {
			return fmt.Errorf("%w while preparing data for indexing", errP)
		}

		for i := 0; i < len(dataBuffers); i++ {
			err := r.destinationElastic.DoBulkRequest(dataBuffers[i], index)
			if err != nil {
				return fmt.Errorf("%w while r.destinationElastic.DoBulkRequest", err)
			}
		}
		return nil
	}
}
