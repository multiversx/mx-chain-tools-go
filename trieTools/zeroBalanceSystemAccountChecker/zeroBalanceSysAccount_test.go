package main

import (
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExportSystemAccZeroTokensBalances(t *testing.T) {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	require.Nil(t, err)

	systemSCAddress := addressConverter.Encode(vmcommon.SystemAccountAddress)
	globalTokens := make(map[string]map[string]struct{})
	globalTokens["adr1"] = map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
	}
	globalTokens["adr2"] = map[string]struct{}{
		"token3-r-0": {},
	}
	globalTokens[systemSCAddress] = map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
		"token3-r-0": {},
		"token4-r-0": {},
		"token5-r-0": {},
	}

	shardAddressTokenMap := make(map[uint32]map[string]map[string]struct{})

	// Shard 0
	shardAddressTokenMap[0] = make(map[string]map[string]struct{})
	shardAddressTokenMap[0][systemSCAddress] = map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
		"token4-r-0": {},
	}
	shardAddressTokenMap[0]["adr1"] = map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
	}

	// Shard 1
	shardAddressTokenMap[1] = make(map[string]map[string]struct{})
	shardAddressTokenMap[1][systemSCAddress] = map[string]struct{}{
		"token3-r-0": {},
		"token5-r-0": {},
	}
	shardAddressTokenMap[1]["adr2"] = map[string]struct{}{
		"token3-r-0": {},
	}

	globalExtraTokens, shardExtraTokens, err := exportSystemAccZeroTokensBalances(globalTokens, shardAddressTokenMap)
	require.Nil(t, err)
	expectedShardExtraTokens := make(map[uint32]map[string]struct{})
	expectedShardExtraTokens[0] = map[string]struct{}{
		"token4-r-0": {},
	}
	expectedShardExtraTokens[1] = map[string]struct{}{
		"token5-r-0": {},
	}
	require.Equal(t, expectedShardExtraTokens, shardExtraTokens)

	expectedGlobalExtraTokens := map[string]struct{}{
		"token4-r-0": {},
		"token5-r-0": {},
	}
	require.Equal(t, expectedGlobalExtraTokens, globalExtraTokens)
}
