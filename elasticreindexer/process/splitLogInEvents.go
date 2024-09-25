package process

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

type LogsResponse struct {
	Hits struct {
		Hits []struct {
			ID     string     `json:"_id"`
			Source *data.Logs `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type logsToEvensProcessor struct {
	addressConverter core.PubkeyConverter
}

func NewLogsToEvensProcessor() (*logsToEvensProcessor, error) {
	addressConverter, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	if err != nil {
		return nil, err
	}

	return &logsToEvensProcessor{
		addressConverter: addressConverter,
	}, nil
}

func (lep *logsToEvensProcessor) Process(responseBytes []byte) ([]*bytes.Buffer, uint64, error) {
	logsResponse := &LogsResponse{}
	err := json.Unmarshal(responseBytes, logsResponse)
	if err != nil {
		return nil, 0, err
	}

	logEvents := make([]*data.LogEvent, 0)
	for _, dbLog := range logsResponse.Hits.Hits {
		dbLogEvents, errC := lep.createEventsFromLog(dbLog.ID, dbLog.Source)
		if errC != nil {
			return nil, 0, errC
		}

		logEvents = append(logEvents, dbLogEvents...)
	}

	buffSlice := data.NewBufferSlice(0)
	for _, dbLog := range logEvents {
		err = serializeLogEvent(dbLog, buffSlice)
		if err != nil {
			return nil, 0, err
		}
	}

	return buffSlice.Buffers(), uint64(len(logEvents)), nil
}

func (lep *logsToEvensProcessor) createEventsFromLog(txHash string, log *data.Logs) ([]*data.LogEvent, error) {
	logEvents := make([]*data.LogEvent, 0, len(log.Events))
	for _, event := range log.Events {
		addressBytes, err := lep.addressConverter.Decode(log.Address)
		if err != nil {
			return nil, err
		}

		eventShardID := sharding.ComputeShardID(addressBytes, 3)
		logEvents = append(logEvents, &data.LogEvent{
			ID:             fmt.Sprintf("%s-%d-%d", txHash, eventShardID, event.Order),
			TxHash:         txHash,
			OriginalTxHash: log.OriginalTxHash,
			LogAddress:     log.Address,
			Address:        event.Address,
			Identifier:     event.Identifier,
			Data:           hex.EncodeToString(event.Data),
			AdditionalData: hexEncodeSlice(event.AdditionalData),
			Topics:         hexEncodeSlice(event.Topics),
			Order:          event.Order,
			ShardID:        eventShardID,
			Timestamp:      log.Timestamp,
		})
	}

	return logEvents, nil
}

func hexEncodeSlice(input [][]byte) []string {
	hexEncoded := make([]string, 0, len(input))
	for idx := 0; idx < len(input); idx++ {
		hexEncoded = append(hexEncoded, hex.EncodeToString(input[idx]))
	}
	if len(hexEncoded) == 0 {
		return nil
	}

	return hexEncoded
}

func serializeLogEvent(logEvent *data.LogEvent, buffSlice *data.BufferSlice) error {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, logEvent.ID, "\n"))
	serializedData, errMarshal := json.Marshal(logEvent)
	if errMarshal != nil {
		return errMarshal
	}

	return buffSlice.PutData(meta, serializedData)
}
