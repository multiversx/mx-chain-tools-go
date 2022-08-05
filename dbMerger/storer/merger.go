package storer

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("storer")

// MergeDBs will iterate over all provided sources and take all key-value pairs and write them in the destination persister
func MergeDBs(dest storage.Persister, sources ...storage.Persister) error {
	err := checkArgs(dest, sources...)
	if err != nil {
		return err
	}

	numKeys := 0

	for _, source := range sources {
		copiedKeys, errMerge := mergeDB(dest, source)
		if errMerge != nil {
			return errMerge
		}

		numKeys += copiedKeys
	}

	log.Debug("finished copying data",
		"num source persisters", len(sources), "num key-values copied", numKeys)

	return nil
}

func checkArgs(dest storage.Persister, sources ...storage.Persister) error {
	if check.IfNil(dest) {
		return fmt.Errorf("%w for the destination persister", errNilPersister)
	}
	for idx, source := range sources {
		if check.IfNil(source) {
			return fmt.Errorf("%w for the source persister, index %d", errNilPersister, idx)
		}
	}

	return nil
}

func mergeDB(dest storage.Persister, source storage.Persister) (int, error) {
	var foundErr error
	numKeysCopied := 0
	source.RangeKeys(func(key []byte, val []byte) bool {
		numKeysCopied++
		foundErr = dest.Put(key, val)
		if foundErr != nil {
			return false
		}

		return true
	})

	return numKeysCopied, foundErr
}
