package elasticsearch

import (
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"

	"quorumengineering/quorum-report/types"
)

func NewClient(config elasticsearch7.Config) (*elasticsearch7.Client, error) {
	return elasticsearch7.NewClient(config)
}

func NewConfig(config *types.ElasticSearchConfig) elasticsearch7.Config {
	return elasticsearch7.Config{
		Addresses: config.Addresses,
		CloudID:   config.CloudID,
	}
}
