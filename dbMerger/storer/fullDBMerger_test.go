package storer

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-tools-go/dbmerger/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArgsFullDBMerger() ArgsFullDBMerger {
	return ArgsFullDBMerger{
		DataMergerInstance: &mock.DataMergerStub{},
		PersisterCreator:   &mock.PersisterCreatorStub{},
		CopyHandler:        &mock.CopyHandlerStub{},
	}
}

func TestNewFullDBMerger(t *testing.T) {
	t.Parallel()

	t.Run("nil DataMergerInstance", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFullDBMerger()
		args.DataMergerInstance = nil
		merger, err := NewFullDBMerger(args)

		assert.True(t, check.IfNil(merger))
		assert.True(t, errors.Is(err, errNilComponent))
		assert.True(t, strings.Contains(err.Error(), "DataMergerInstance"))
	})
	t.Run("nil PersisterCreator", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFullDBMerger()
		args.PersisterCreator = nil
		merger, err := NewFullDBMerger(args)

		assert.True(t, check.IfNil(merger))
		assert.True(t, errors.Is(err, errNilComponent))
		assert.True(t, strings.Contains(err.Error(), "PersisterCreator"))
	})
	t.Run("nil CopyHandler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFullDBMerger()
		args.CopyHandler = nil
		merger, err := NewFullDBMerger(args)

		assert.True(t, check.IfNil(merger))
		assert.True(t, errors.Is(err, errNilComponent))
		assert.True(t, strings.Contains(err.Error(), "CopyHandler"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFullDBMerger()
		merger, err := NewFullDBMerger(args)

		assert.False(t, check.IfNil(merger))
		assert.Nil(t, err)
	})
}

func TestDataMerger_MergeDBs(t *testing.T) {
	t.Parallel()

	t.Run("invalid number of source paths", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFullDBMerger()
		merger, _ := NewFullDBMerger(args)

		destPersister, err := merger.MergeDBs("dest")
		assert.True(t, check.IfNil(destPersister))
		assert.True(t, errors.Is(err, errInvalidNumberOfPersisters))
		assert.True(t, strings.Contains(err.Error(), "provided 0, minimum 2"))

		destPersister, err = merger.MergeDBs("dest", "src1")
		assert.True(t, check.IfNil(destPersister))
		assert.True(t, errors.Is(err, errInvalidNumberOfPersisters))
		assert.True(t, strings.Contains(err.Error(), "provided 1, minimum 2"))
	})
	t.Run("directory copy errors", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsFullDBMerger()
		args.CopyHandler = &mock.CopyHandlerStub{
			CopyDirectoryCalled: func(destination string, source string) error {
				return expectedErr
			},
		}
		merger, _ := NewFullDBMerger(args)

		destPersister, err := merger.MergeDBs("dest", "src1", "src2")
		assert.True(t, check.IfNil(destPersister))
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("create destination persister errors", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsFullDBMerger()
		args.PersisterCreator = &mock.PersisterCreatorStub{
			CreatePersisterCalled: func(path string) (storage.Persister, error) {
				if path == "dest" {
					return nil, expectedErr
				}

				return mock.NewPersisterMock(), nil
			},
		}
		merger, _ := NewFullDBMerger(args)

		destPersister, err := merger.MergeDBs("dest", "src1", "src2")
		assert.True(t, check.IfNil(destPersister))
		assert.True(t, errors.Is(err, expectedErr))
		assert.True(t, strings.Contains(err.Error(), "for destination persister"))
	})
	t.Run("create source persister errors", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsFullDBMerger()
		args.PersisterCreator = &mock.PersisterCreatorStub{
			CreatePersisterCalled: func(path string) (storage.Persister, error) {
				if strings.Contains(path, "src") {
					return nil, expectedErr
				}

				return mock.NewPersisterMock(), nil
			},
		}
		merger, _ := NewFullDBMerger(args)

		destPersister, err := merger.MergeDBs("dest", "src1", "src2")
		assert.True(t, check.IfNil(destPersister))
		assert.True(t, errors.Is(err, expectedErr))
		assert.True(t, strings.Contains(err.Error(), "for source persister with index 1"))
	})
	t.Run("data merge errors", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsFullDBMerger()
		args.PersisterCreator = &mock.PersisterCreatorStub{
			CreatePersisterCalled: func(path string) (storage.Persister, error) {
				return mock.NewPersisterMock(), nil
			},
		}
		args.DataMergerInstance = &mock.DataMergerStub{
			MergeDBsCalled: func(dest storage.Persister, sources ...storage.Persister) error {
				return expectedErr
			},
		}
		merger, _ := NewFullDBMerger(args)

		destPersister, err := merger.MergeDBs("dest", "src1", "src2", "src3")
		assert.True(t, check.IfNil(destPersister))
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		numClosedPersisters := 0
		copyCalled := false
		numPersistersCreated := 0
		mergeDBCalled := false
		args := createMockArgsFullDBMerger()
		args.CopyHandler = &mock.CopyHandlerStub{
			CopyDirectoryCalled: func(destination string, source string) error {
				copyCalled = true

				return nil
			},
		}
		args.PersisterCreator = &mock.PersisterCreatorStub{
			CreatePersisterCalled: func(path string) (storage.Persister, error) {
				numPersistersCreated++
				persisterMock := mock.NewPersisterMock()
				persisterMock.CloseCalled = func() error {
					numClosedPersisters++

					return nil
				}
				return persisterMock, nil
			},
		}
		args.DataMergerInstance = &mock.DataMergerStub{
			MergeDBsCalled: func(dest storage.Persister, sources ...storage.Persister) error {
				assert.Equal(t, 2, len(sources))
				assert.False(t, check.IfNil(dest))
				mergeDBCalled = true

				return nil
			},
		}
		merger, _ := NewFullDBMerger(args)

		destPersister, err := merger.MergeDBs("dest", "src1", "src2", "src3")
		assert.False(t, check.IfNil(destPersister))
		assert.Nil(t, err)
		assert.True(t, copyCalled)
		assert.Equal(t, 3, numPersistersCreated)
		assert.True(t, mergeDBCalled)
		assert.Equal(t, 2, numClosedPersisters) // 3 sources, 1 copied, 2 opened to copy key by key
	})
}
