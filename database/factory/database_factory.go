package factory

import (
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/database/elasticsearch"
	"quorumengineering/quorum-report/database/memory"
	"quorumengineering/quorum-report/types"
)

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (dbFactory *Factory) Database(config *types.DatabaseConfig) (database.Database, error) {
	if config != nil && config.Elasticsearch != nil {
		db, err := dbFactory.NewElasticsearchDatabase(config.Elasticsearch)
		if err != nil {
			return nil, err
		}
		return NewDatabaseWithCache(db, config.CacheSize)
	}
	return dbFactory.NewInMemoryDatabase(), nil
}

func (dbFactory *Factory) NewInMemoryDatabase() *memory.MemoryDB {
	return memory.NewMemoryDB()
}

func (dbFactory *Factory) NewElasticsearchDatabase(config *types.ElasticsearchConfig) (*elasticsearch.ElasticsearchDB, error) {
	esConfig := elasticsearch.NewConfig(config)
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, err
	}
	return elasticsearch.New(elasticsearch.NewAPIClient(client))
}
