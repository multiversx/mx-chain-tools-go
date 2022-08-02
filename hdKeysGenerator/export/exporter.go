package export

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-tools-go/hdKeysGenerator/common"
)

// ArgsNewExporter holds arguments for creating an exporter
type ArgsNewExporter struct {
	ActualShard    common.OptionalUint32
	ProjectedShard common.OptionalUint32
	StartIndex     int
	NumKeys        int
	Format         string
}

type exporter struct {
	actualShard    common.OptionalUint32
	projectedShard common.OptionalUint32
	startIndex     int
	numKeys        int
	format         string
}

// NewExporter creates a new exporter
func NewExporter(args ArgsNewExporter) (*exporter, error) {
	return &exporter{
		actualShard:    args.ActualShard,
		projectedShard: args.ProjectedShard,
		startIndex:     args.StartIndex,
		numKeys:        args.NumKeys,
		format:         args.Format,
	}, nil
}

func (e *exporter) ExportKeys(keys []common.GeneratedKey) error {
	log.Info("Exporting:",
		"numKeys", len(keys),
		"formatType", e.format,
	)

	err := e.saveKeysFile(keys)
	if err != nil {
		return err
	}

	err = e.saveMetadataFile()
	if err != nil {
		return err
	}

	return nil
}

func (e *exporter) saveKeysFile(keys []common.GeneratedKey) error {
	formatter, err := e.getFormatter()
	if err != nil {
		return err
	}

	text, err := formatter.toText(keys)
	if err != nil {
		return err
	}

	fileBasename := e.getOutputFileBasename()
	filename := fmt.Sprintf("%s.%s", fileBasename, formatter.getFileExtension())
	err = e.saveFile(filename, text)
	if err != nil {
		return err
	}

	return nil
}

func (e *exporter) getFormatter() (formatter, error) {
	switch e.format {
	case FormatterNamePlainText:
		return &formatterPlainText{}, nil
	case FormatterNamePlainJson:
		return &formatterPlainJson{}, nil
	}

	return nil, fmt.Errorf("unknown format: %s", e.format)
}

func (e *exporter) getOutputFileBasename() string {
	actualShardPart := "actualShard_None"
	projectedShardPart := "projectedShard_None"
	startIndexPart := fmt.Sprintf("start_%d", e.startIndex)
	numKeysPart := fmt.Sprintf("numKeys_%d", e.numKeys)

	if e.actualShard.HasValue {
		actualShardPart = fmt.Sprintf("ActualShard_%d", e.actualShard.Value)
	}
	if e.projectedShard.HasValue {
		projectedShardPart = fmt.Sprintf("ProjectedShard_%d", e.projectedShard.Value)
	}

	return fmt.Sprintf("Keys_%s_%s_%s_%s", actualShardPart, projectedShardPart, startIndexPart, numKeysPart)
}

func (e *exporter) saveMetadataFile() error {
	metadata := &exportMetadata{
		ActualShardID: 0,
	}

	metadataJson, err := json.MarshalIndent(metadata, "", fourSpaces)
	if err != nil {
		return err
	}

	fileBasename := e.getOutputFileBasename()
	metadataFilename := fmt.Sprintf("%s.%s.metadata.json", fileBasename, e.format)
	err = e.saveFile(metadataFilename, string(metadataJson))
	if err != nil {
		return err
	}

	return nil
}

func (e *exporter) saveFile(filename string, text string) error {
	err := ioutil.WriteFile(filename, []byte(text), core.FileModeReadWrite)
	if err != nil {
		return err
	}

	log.Info("Saved file:", "file", filename)
	return nil
}
