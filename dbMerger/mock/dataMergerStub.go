package mock

import "github.com/ElrondNetwork/elrond-go/storage"

// DataMergerStub -
type DataMergerStub struct {
	MergeDBsCalled func(dest storage.Persister, sources ...storage.Persister) error
}

// MergeDBs -
func (stub *DataMergerStub) MergeDBs(dest storage.Persister, sources ...storage.Persister) error {
	if stub.MergeDBsCalled != nil {
		return stub.MergeDBsCalled(dest, sources...)
	}

	return nil
}

// IsInterfaceNil -
func (stub *DataMergerStub) IsInterfaceNil() bool {
	return stub == nil
}
