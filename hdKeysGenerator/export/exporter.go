package export

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

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
	OutputFile     string
}

type exporter struct {
	actualShard    common.OptionalUint32
	projectedShard common.OptionalUint32
	startIndex     int
	numKeys        int
	format         string
	outputFile     string
}

// NewExporter creates a new exporter
func NewExporter(args ArgsNewExporter) (*exporter, error) {
	return &exporter{
		actualShard:    args.ActualShard,
		projectedShard: args.ProjectedShard,
		startIndex:     args.StartIndex,
		numKeys:        args.NumKeys,
		format:         args.Format,
		outputFile:     args.OutputFile,
	}, nil
}

// ExportKeys exports the generated keys to a file
func (e *exporter) ExportKeys(keys []common.GeneratedKey) error {
	log.Info("exporting:",
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

	err = e.saveFile(e.outputFile, text)
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
	return strings.TrimSuffix(e.outputFile, filepath.Ext(e.outputFile))
}

func (e *exporter) saveMetadataFile() error {
	metadata := &exportMetadata{
		ActualShardID:          e.actualShard.Value,
		ActualShardHasValue:    e.actualShard.HasValue,
		ProjectedShardID:       e.projectedShard.Value,
		ProjectedShardHasValue: e.projectedShard.HasValue,
		StartIndex:             e.startIndex,
		NumKeys:                e.numKeys,
	}

	metadataJson, err := json.MarshalIndent(metadata, "", fourSpaces)
	if err != nil {
		return err
	}

	fileBasename := strings.TrimSuffix(e.outputFile, filepath.Ext(e.outputFile))
	metadataFilename := fmt.Sprintf("%s.metadata.json", fileBasename)
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

	log.Info("saved file:", "file", filename)
	return nil
}
