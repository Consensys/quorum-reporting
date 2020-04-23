package factory

import (
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/database/elasticsearch"
	"quorumengineering/quorum-report/database/memory"
	"quorumengineering/quorum-report/types"
)

type Factory struct {
}

func NewFactory() *Factory {
	return &Factory{}
}

func (dbFactory *Factory) Database(config types.ReportInputStruct) (database.Database, error) {
	if config.ElasticSearch != nil {
		return dbFactory.NewElasticSearchDatabase(config.ElasticSearch)
	}
	return dbFactory.NewInMemoryDatabase(), nil
}

func (dbFactory *Factory) NewInMemoryDatabase() *memory.MemoryDB {
	return memory.NewMemoryDB()
}

func (dbFactory *Factory) NewElasticSearchDatabase(config *types.ElasticSearchConfig) (*elasticsearch.ElasticsearchDB, error) {
	esConfig := elasticsearch.NewConfig(config)
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, err
	}
	return elasticsearch.New(elasticsearch.NewAPIClient(client)), nil
}
