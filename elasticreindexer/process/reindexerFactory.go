package process

import (
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/config"
	"github.com/ElrondNetwork/elrond-tools-go/elasticreindexer/elastic"
)

// CreateReindexer will create the source and destination elastic handlers and create a reindexer based on them
func CreateReindexer(cfg *config.GeneralConfig) (*reindexer, error) {
	sourceElastic, err := elastic.NewElasticClient(cfg.Indexers.Input)
	if err != nil {
		return nil, err
	}

	destinationElastic, err := elastic.NewElasticClient(cfg.Indexers.Output)
	if err != nil {
		return nil, err
	}

	return newReindexer(sourceElastic, destinationElastic, cfg.Indexers.IndicesConfig.Indices)
}
