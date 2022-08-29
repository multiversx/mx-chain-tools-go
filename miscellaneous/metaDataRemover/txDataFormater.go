package main

import (
	"bytes"
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
)

func createTxsData(tokens map[string][]*interval, intervalBulkSize uint64) ([][]byte, error) {
	tokensData := tokensMapToOrderedArray(tokens)

	txsData := make([][]byte, 0)
	numTokensInBulk := uint64(0)
	intervalsInBulk := make([]*interval, 0, intervalBulkSize)
	txDataBuilder := builders.NewTxDataBuilder().Function(ESDTDeleteMetadataPrefix)
	for _, tkData := range tokensData {
		tokenIDHex := hex.EncodeToString([]byte(tkData.tokenID))
		txDataBuilder.ArgHexString(tokenIDHex)

		intervalsCopy := make([]*interval, len(tkData.intervals))
		copy(intervalsCopy, tkData.intervals)

		intervalIndex := 0
		for intervalIndex < len(intervalsCopy) {
			currInterval := intervalsCopy[intervalIndex]

			tokensInInterval := currInterval.end - currInterval.start + 1
			availableSlots := intervalBulkSize - numTokensInBulk
			if availableSlots >= tokensInInterval {
				intervalsInBulk = append(intervalsInBulk, currInterval)
				numTokensInBulk += tokensInInterval
			} else {
				first, second := splitInterval(currInterval, availableSlots)

				intervalsCopy = append(intervalsCopy, second)
				intervalsInBulk = append(intervalsInBulk, first)
				numTokensInBulk += availableSlots
			}

			bulkFull := numTokensInBulk == intervalBulkSize
			lastInterval := intervalIndex == len(intervalsCopy)-1
			shouldEmptyBulk := lastInterval && numTokensInBulk != 0
			if bulkFull || shouldEmptyBulk {
				addIntervalsAsOnData(txDataBuilder, intervalsInBulk)
				intervalsInBulk = make([]*interval, 0, intervalBulkSize)
			}

			if bulkFull {
				currTxData, err := txDataBuilder.ToDataBytes()
				if err != nil {
					return nil, err
				}

				numTokensInBulk = 0
				txsData = append(txsData, currTxData)
				txDataBuilder = builders.NewTxDataBuilder().Function(ESDTDeleteMetadataPrefix).ArgHexString(tokenIDHex)

				if lastInterval {
					txDataBuilder = builders.NewTxDataBuilder().Function(ESDTDeleteMetadataPrefix)
				}
			}

			intervalIndex++
		}
	}

	remainingTxData, err := getRemainingTxData(txDataBuilder)
	if err != nil {
		return nil, err
	}
	if len(remainingTxData) > 0 {
		txsData = append(txsData, remainingTxData)
	}

	return txsData, nil
}

/*
func createTxsData2(bulkTokens [][]*tokenWithInterval) {
	txDataBuilder := builders.NewTxDataBuilder().Function(ESDTDeleteMetadataPrefix)

	for _, bulk := range bulkTokens {

	}
}

func addTokensAsOnData(txDataBuilder builders.TxDataBuilder, tokens []*tokenWithInterval){
	for _, tkData := range tokens{
		tokenIDHex := hex.EncodeToString([]byte(tkData.tokenID))
		txDataBuilder.ArgHexString(tokenIDHex)
		addIntervalsAsOnData(txDataBuilder, tkData.interval)
	}
}
*/

func splitInterval(currInterval *interval, index uint64) (*interval, *interval) {
	first := &interval{
		start: currInterval.start,
		end:   currInterval.start + index - 1,
	}

	second := &interval{
		start: first.end + 1,
		end:   currInterval.end,
	}

	return first, second
}

func addIntervalsAsOnData(builder builders.TxDataBuilder, intervals []*interval) {
	builder.ArgInt64(int64(len(intervals)))

	for _, interval := range intervals {
		builder.
			ArgInt64(int64(interval.start)).
			ArgInt64(int64(interval.end))
	}
}

func getRemainingTxData(txDataBuilder builders.TxDataBuilder) ([]byte, error) {
	txData, err := txDataBuilder.ToDataBytes()
	if err != nil {
		return nil, err
	}

	splits := bytes.Split(txData, []byte("@"))
	if len(splits) > 2 {
		return txData, nil
	}

	return nil, nil
}
