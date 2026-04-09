package main

import "github.com/multiversx/mx-chain-go/common"

type LastSnapshotMarkerStub struct {
	AddMarkerCalled     func(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte)
	RemoveMarkerCalled  func(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte)
	GetMarkerInfoCalled func(trieStorageManager common.StorageManager) ([]byte, error)
}

func (lsm *LastSnapshotMarkerStub) AddMarker(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte) {
	if lsm.AddMarkerCalled != nil {
		lsm.AddMarkerCalled(trieStorageManager, epoch, rootHash)
	}
}

func (lsm *LastSnapshotMarkerStub) RemoveMarker(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte) {
	if lsm.RemoveMarkerCalled != nil {
		lsm.RemoveMarkerCalled(trieStorageManager, epoch, rootHash)
	}
}

func (lsm *LastSnapshotMarkerStub) GetMarkerInfo(trieStorageManager common.StorageManager) ([]byte, error) {
	if lsm.GetMarkerInfoCalled != nil {
		return lsm.GetMarkerInfoCalled(trieStorageManager)
	}
	return nil, nil
}

func (lsm *LastSnapshotMarkerStub) IsInterfaceNil() bool {
	return lsm == nil
}
