package main

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
)

const ESDTDeleteMetadataPrefix = "ESDTDeleteMetadata"

func createTxsData(bulks [][]*tokenData) ([][]byte, error) {
	txsData := make([][]byte, 0, len(bulks))
	for _, bulk := range bulks {
		txData, err := tokensBulkAsOnData(bulk)
		if err != nil {
			return nil, err
		}

		txsData = append(txsData, txData)
	}

	return txsData, nil
}

func tokensBulkAsOnData(bulk []*tokenData) ([]byte, error) {
	txDataBuilder := builders.NewTxDataBuilder().Function(ESDTDeleteMetadataPrefix)
	for _, tkData := range bulk {
		tokenIDHex := hex.EncodeToString([]byte(tkData.tokenID))
		txDataBuilder.ArgHexString(tokenIDHex)

		addIntervalsAsOnData(txDataBuilder, tkData.intervals)
	}

	return txDataBuilder.ToDataBytes()
}

func addIntervalsAsOnData(builder builders.TxDataBuilder, intervals []*interval) {
	builder.ArgInt64(int64(len(intervals)))

	for _, interval := range intervals {
		builder.
			ArgInt64(int64(interval.start)).
			ArgInt64(int64(interval.end))
	}
}
