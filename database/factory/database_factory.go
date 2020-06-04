package factory

import (
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/database/elasticsearch"
	"quorumengineering/quorum-report/database/memory"
	"quorumengineering/quorum-report/log"
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
		log.Info("created database connection", "type", "elasticsearch")
		return NewDatabaseWithCache(db, config.CacheSize)
	}
	log.Info("created database connection", "type", "memory")
	return dbFactory.NewInMemoryDatabase(), nil
}

func (dbFactory *Factory) NewInMemoryDatabase() *memory.MemoryDB {
	return memory.NewMemoryDB()
}

func (dbFactory *Factory) NewElasticsearchDatabase(config *types.ElasticsearchConfig) (*elasticsearch.ElasticsearchDB, error) {
	esConfig, err := elasticsearch.NewConfig(config)
	if err != nil {
		return nil, err
	}
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, err
	}
	apiClient, err := elasticsearch.NewAPIClient(client)
	if err != nil {
		return nil, err
	}
	return elasticsearch.New(apiClient)
}
