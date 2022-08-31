package main

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-tools-go/trieTools/trieToolsCommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestExportSystemAccZeroTokensBalances(t *testing.T) {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	require.Nil(t, err)

	systemSCAddress := addressConverter.Encode(vmcommon.SystemAccountAddress)
	globalTokens := trieToolsCommon.NewAddressTokensMap()
	globalTokens.Add("adr1", map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
	})
	globalTokens.Add("adr2", map[string]struct{}{
		"token3-r-0": {},
	})
	globalTokens.Add(systemSCAddress, map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
		"token3-r-0": {},
		"token4-r-0": {},
		"token5-r-0": {},
	})

	shardAddressTokenMap := make(map[uint32]trieToolsCommon.AddressTokensMap)

	// Shard 0
	shardAddressTokenMap[0] = trieToolsCommon.NewAddressTokensMap()
	shardAddressTokenMap[0].Add(systemSCAddress, map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
		"token4-r-0": {},
	})
	shardAddressTokenMap[0].Add("adr1", map[string]struct{}{
		"token1-r-0": {},
		"token2-r-0": {},
	})

	// Shard 1
	shardAddressTokenMap[1] = trieToolsCommon.NewAddressTokensMap()
	shardAddressTokenMap[1].Add(systemSCAddress, map[string]struct{}{
		"token3-r-0": {},
		"token5-r-0": {},
	})
	shardAddressTokenMap[1].Add("adr2", map[string]struct{}{
		"token3-r-0": {},
	})

	exporter, err := newZeroTokensBalancesExporter(addressConverter)
	require.Nil(t, err)
	globalExtraTokens, shardExtraTokens, err := exporter.getExtraTokens(globalTokens, shardAddressTokenMap)
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
