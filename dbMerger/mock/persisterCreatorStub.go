package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/storage"
)

// PersisterCreatorStub -
type PersisterCreatorStub struct {
	CreatePersisterCalled func(path string) (storage.Persister, error)
}

// CreatePersister -
func (stub *PersisterCreatorStub) CreatePersister(path string) (storage.Persister, error) {
	if stub.CreatePersisterCalled != nil {
		return stub.CreatePersisterCalled(path)
	}

	return nil, errors.New("not implemented")
}

// IsInterfaceNil -
func (stub *PersisterCreatorStub) IsInterfaceNil() bool {
	return stub == nil
}
