package process

import (
	"github.com/ElrondNetwork/elrond-tools-go/sccounter/config"
	"github.com/ElrondNetwork/elrond-tools-go/sccounter/elastic"
)

func CreateSCCounter(cfg *config.GeneralConfig) (*scCounter, error) {
	elasticClient, err := elastic.NewElasticClient(cfg.SCDeploysConfig.ElasticInstance)
	if err != nil {
		return nil, err
	}

	return newSCCounter(elasticClient)
}
