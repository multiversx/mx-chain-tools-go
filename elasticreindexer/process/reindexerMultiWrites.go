package process

import (
	"errors"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
	"sync"
	"time"
)

type interval struct {
	start int64
	stop  int64
}

type reindexerMultiWrite struct {
	indicesNoTimestamp   []string
	indicesWithTimestamp []string
	numParallelWrite     int
	blockChainStartTime  int64

	reindexerClient *reindexer
}

func NewReindexerMultiWrite(cfg *config.GeneralConfig) (*reindexerMultiWrite, error) {
	ri, err := CreateReindexer(cfg)
	if err != nil {
		return nil, err
	}
	return &reindexerMultiWrite{
		reindexerClient:      ri,
		indicesNoTimestamp:   cfg.Indexers.IndicesConfig.Indices,
		indicesWithTimestamp: cfg.Indexers.IndicesConfig.WithTimestamp.IndicesWithTimestamp,
		numParallelWrite:     cfg.Indexers.IndicesConfig.WithTimestamp.NumParallelWrites,
		blockChainStartTime:  cfg.Indexers.IndicesConfig.WithTimestamp.BlockchainStartTime,
	}, nil
}

func (rmw *reindexerMultiWrite) ProcessNoTimestamp(overwrite bool, skipMappings bool) error {
	for _, index := range rmw.indicesNoTimestamp {
		err := rmw.reindexerClient.Process(overwrite, skipMappings, index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rmw *reindexerMultiWrite) ProcessWithTimestamp(overwrite bool, skipMappings bool) error {
	intervals, err := computeIntervals(rmw.blockChainStartTime, int64(rmw.numParallelWrite))
	if err != nil {
		return err
	}

	for _, index := range rmw.indicesWithTimestamp {
		wg := &sync.WaitGroup{}
		wg.Add(rmw.numParallelWrite)

		log.Info("starting reindexing", "index", index)

		count := uint64(0)

		for _, interv := range intervals {
			go func(startTime, stopTime int64) {
				defer wg.Done()
				errIndex := rmw.processIndexWithTimestamp(index, overwrite, skipMappings, startTime, stopTime, &count)
				if errIndex != nil {
					log.Warn("rmw.processIndexWithTimestamp", "index", index, "error", errIndex.Error())
				}
				time.Sleep(time.Second)
			}(interv.start, interv.stop)
		}

		wg.Wait()
	}

	return nil
}

func (rmw *reindexerMultiWrite) processIndexWithTimestamp(index string, overwrite bool, skipMappings bool, start, stop int64, count *uint64) error {
	return rmw.reindexerClient.processIndexWithTimestamp(index, overwrite, skipMappings, start, stop, count)
}

func computeIntervals(blockchainStartTime int64, numIntervals int64) ([]*interval, error) {
	currentTimestampUnix := time.Now().Unix()
	startTime := blockchainStartTime

	if startTime > currentTimestampUnix {
		return nil, errors.New("blockchain start time is greater than current timestamp")
	}
	if numIntervals < 2 {
		return []*interval{{
			start: blockchainStartTime,
			stop:  currentTimestampUnix,
		}}, nil
	}

	difference := currentTimestampUnix - startTime

	step := difference / numIntervals

	intervals := make([]*interval, 0)
	for idx := int64(0); idx < numIntervals; idx++ {
		intervals = append(intervals, &interval{
			start: startTime + idx*step,
			stop:  startTime + (idx+1)*step,
		})
	}

	return intervals, nil
}
