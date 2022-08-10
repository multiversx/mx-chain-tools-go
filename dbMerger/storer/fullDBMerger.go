package storer

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const minNumOfPersisters = 2

// ArgsFullDBMerger is the DTO used in the NewFullDBMerger constructor function
type ArgsFullDBMerger struct {
	DataMergerInstance DataMerger
	PersisterCreator   PersisterCreator
	CopyHandler        CopyHandler
}

type fullDBMerger struct {
	dataMergerInstance DataMerger
	persisterCreator   PersisterCreator
	copyHandler        CopyHandler
}

// NewFullDBMerger creates a new instance of type fullDBMerger
func NewFullDBMerger(args ArgsFullDBMerger) (*fullDBMerger, error) {
	if check.IfNil(args.DataMergerInstance) {
		return nil, fmt.Errorf("%w, DataMergerInstance", errNilComponent)
	}
	if check.IfNil(args.PersisterCreator) {
		return nil, fmt.Errorf("%w, PersisterCreator", errNilComponent)
	}
	if check.IfNil(args.CopyHandler) {
		return nil, fmt.Errorf("%w, CopyHandler", errNilComponent)
	}

	return &fullDBMerger{
		dataMergerInstance: args.DataMergerInstance,
		persisterCreator:   args.PersisterCreator,
		copyHandler:        args.CopyHandler,
	}, nil
}

// MergeDBs will merge all data from the source persiste paths into a new storage persister
func (fdm *fullDBMerger) MergeDBs(destinationPath string, sourcePaths ...string) (storage.Persister, error) {
	if len(sourcePaths) < minNumOfPersisters {
		return nil, fmt.Errorf("%w, provided %d, minimum %d", errInvalidNumberOfPersisters, len(sourcePaths), minNumOfPersisters)
	}

	err := fdm.copyHandler.CopyDirectory(destinationPath, sourcePaths[0])
	if err != nil {
		return nil, err
	}

	destPersister, err := fdm.persisterCreator.CreatePersister(destinationPath)
	if err != nil {
		return nil, fmt.Errorf("%w for destination persister", err)
	}

	sourcePersisters, err := fdm.createSourcePersisters(sourcePaths...)
	if err != nil {
		return nil, err
	}

	err = fdm.dataMergerInstance.MergeDBs(destPersister, sourcePersisters...)
	if err != nil {
		return nil, err
	}

	return destPersister, nil
}

func (fdm *fullDBMerger) createSourcePersisters(sourcePaths ...string) ([]storage.Persister, error) {
	sourcePersisters := make([]storage.Persister, 0, len(sourcePaths)-1)
	for i := 1; i < len(sourcePaths); i++ {
		srcPersister, errPersister := fdm.persisterCreator.CreatePersister(sourcePaths[i])
		if errPersister != nil {
			return nil, fmt.Errorf("%w for source persister with index %d", errPersister, i)
		}

		sourcePersisters = append(sourcePersisters, srcPersister)
	}

	return sourcePersisters, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (fdm *fullDBMerger) IsInterfaceNil() bool {
	return fdm == nil
}
