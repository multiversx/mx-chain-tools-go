package main

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type constraints struct {
	actualShard               optionalUint32
	projectedShard            optionalUint32
	actualShardCoordinator    sharding.Coordinator
	projectedShardCoordinator sharding.Coordinator
}

func newConstraints(numShards uint32, actualShard optionalUint32, projectedShard optionalUint32) (*constraints, error) {
	actualShardCoordinator, err := sharding.NewMultiShardCoordinator(numShards, actualShard.Value)
	if err != nil {
		return nil, err
	}

	projectedShardCoordinator, err := sharding.NewMultiShardCoordinator(core.MaxNumShards, projectedShard.Value)
	if err != nil {
		return nil, err
	}

	return &constraints{
		actualShard:               actualShard,
		projectedShard:            projectedShard,
		actualShardCoordinator:    actualShardCoordinator,
		projectedShardCoordinator: projectedShardCoordinator,
	}, nil
}

func (c *constraints) areSatisfiedByPublicKey(publicKey []byte) bool {
	if c.projectedShard.HasValue {
		matchesShard := c.projectedShardCoordinator.ComputeId(publicKey) == c.projectedShardCoordinator.SelfId()
		if !matchesShard {
			return false
		}
	}

	if c.actualShard.HasValue {
		matchesShard := c.actualShardCoordinator.ComputeId(publicKey) == c.actualShardCoordinator.SelfId()
		if !matchesShard {
			return false
		}
	}

	return true
}
