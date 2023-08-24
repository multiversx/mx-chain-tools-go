package reader

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// GetElasticTemplatesAndPolicies will return elastic templates and policies
func GetElasticTemplatesAndPolicies(filePath string, indexes []string) (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	indexTemplates := make(map[string]*bytes.Buffer)
	indexPolicies := make(map[string]*bytes.Buffer)
	var err error

	for _, index := range indexes {
		indexTemplates[index], err = getDataFromByIndex(filePath, index)
		if err != nil {
			return nil, nil, err
		}

		indexPolicies[index], err = getDataFromByIndex(path.Join(filePath, "policies"), index)
		if err != nil {
			return nil, nil, err
		}
	}

	return indexTemplates, indexPolicies, nil
}

func getDataFromByIndex(path string, index string) (*bytes.Buffer, error) {
	indexTemplate := &bytes.Buffer{}

	fileName := fmt.Sprintf("%s.json", index)
	filePath := filepath.Join(path, fileName)
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("getDataFromByIndex: %w, path %s, error %s", err, filePath, err.Error())
	}

	indexTemplate.Grow(len(fileBytes))
	_, err = indexTemplate.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("getDataFromByIndex: %w, path %s, error %s", err, filePath, err.Error())
	}

	return indexTemplate, nil
}
