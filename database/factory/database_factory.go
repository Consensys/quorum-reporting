package factory

import (
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/database/elasticsearch"
	"quorumengineering/quorum-report/database/memory"
)

type Factory struct {
}

func NewFactory() *Factory {
	return &Factory{}
}

func (dbFactory *Factory) Database() database.Database {
	return dbFactory.NewInMemoryDatabase()
}

func (dbFactory *Factory) NewInMemoryDatabase() *memory.MemoryDB {
	return memory.NewMemoryDB()
}

func (dbFactory *Factory) NewElasticSearchDatabase() *elasticsearch.ElasticsearchDB {
	return elasticsearch.New()
}
