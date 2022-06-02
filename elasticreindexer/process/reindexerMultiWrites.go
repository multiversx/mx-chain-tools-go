package process

import (
	"errors"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
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
	enabled              bool

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
		enabled:              cfg.Indexers.IndicesConfig.WithTimestamp.Enabled,
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
	if !rmw.enabled {
		return nil
	}

	currentTimestampUnix := time.Now().Unix()
	intervals, err := computeIntervals(rmw.blockChainStartTime, currentTimestampUnix, int64(rmw.numParallelWrite))
	if err != nil {
		return err
	}

	for _, index := range rmw.indicesWithTimestamp {
		err = rmw.reindexBasedOnIntervals(index, intervals, overwrite, skipMappings)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rmw *reindexerMultiWrite) reindexBasedOnIntervals(
	index string,
	intervals []*interval,
	overwrite bool,
	skipMappings bool,
) error {
	wg := &sync.WaitGroup{}
	wg.Add(rmw.numParallelWrite)

	log.Info("starting reindexing", "index", index)

	count := uint64(0)

	for _, interv := range intervals {
		go func(startTime, stopTime int64) {
			defer wg.Done()
			errIndex := rmw.reindexerClient.processIndexWithTimestamp(index, overwrite, skipMappings, startTime, stopTime, &count)
			if errIndex != nil {
				log.Warn("rmw.processIndexWithTimestamp", "index", index, "error", errIndex.Error())
			}
			time.Sleep(time.Second)
		}(interv.start, interv.stop)
	}

	wg.Wait()

	return nil
}

func computeIntervals(startTime, endTime int64, numIntervals int64) ([]*interval, error) {
	if startTime > endTime {
		return nil, errors.New("blockchain start time is greater than current timestamp")
	}
	if numIntervals < 2 {
		return []*interval{{
			start: startTime,
			stop:  endTime,
		}}, nil
	}

	difference := endTime - startTime

	step := difference / numIntervals

	intervals := make([]*interval, 0)
	for idx := int64(0); idx < numIntervals; idx++ {
		start := startTime + idx*step
		stop := startTime + (idx+1)*step

		if idx == numIntervals-1 {
			stop = endTime
		}

		intervals = append(intervals, &interval{
			start: start,
			stop:  stop,
		})
	}

	return intervals, nil
}
