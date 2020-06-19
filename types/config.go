package types

import (
	"os"

	"quorumengineering/quorum-report/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/naoina/toml"
)

type ElasticsearchConfig struct {
	Addresses []string `toml:"urls,omitempty"`
	CloudID   string   `toml:"cloudid"`

	Username string `toml:"username"`
	Password string `toml:"password"`
	APIKey   string `toml:"apikey"`

	// Path to PEM-encoded certificate authorities file
	CACert string `toml:"cacert"`
}

type DatabaseConfig struct {
	Elasticsearch *ElasticsearchConfig `toml:"elasticsearch,omitempty"`
	CacheSize     int                  `toml:"cacheSize,omitempty"`
}

type TuningConfig struct {
	BlockProcessingQueueSize   int `toml:"blockProcessingQueueSize"`
	BlockProcessingFlushPeriod int `toml:"blockProcessingFlushPeriod"`
}

type AddressConfig struct {
	Address      common.Address `toml:"address,omitempty"`
	TemplateName string         `toml:"templateName,omitempty"`
}

type TemplateConfig struct {
	TemplateName  string `toml:"templateName,omitempty"`
	ABI           string `toml:"abi,omitempty"`
	StorageLayout string `toml:"storageLayout,omitempty"`
}

type ReportingConfig struct {
	Title     string
	Addresses []*AddressConfig  `toml:"addresses,omitempty"`
	Templates []*TemplateConfig `toml:"templates,omitempty"`
	Database  *DatabaseConfig   `toml:"database,omitempty"`
	Server    struct {
		RPCAddr     string   `toml:"rpcAddr"`
		RPCCorsList []string `toml:"rpcCorsList,omitempty"`
		RPCVHosts   []string `toml:"rpcvHosts,omitempty"`
		UIPort      int      `toml:"uiPort,omitempty"` // Serve a sample UI if provided
	}
	Connection struct {
		WSUrl             string `toml:"wsUrl"`
		GraphQLUrl        string `toml:"graphQLUrl"`
		ReconnectInterval int    `toml:"reconnectInterval,omitempty"`
		MaxReconnectTries int    `toml:"maxReconnectTries,omitempty"`
	}
	Tuning TuningConfig `toml:"tuning,omitempty"`
}

func (rc *ReportingConfig) SetDefaults() {
	if rc.Tuning.BlockProcessingQueueSize < 1 {
		log.Warn("tuning.BlockProcessingQueueSize below limit", "old value", rc.Tuning.BlockProcessingQueueSize, "new value", 100)
		rc.Tuning.BlockProcessingQueueSize = 100
	}
	if rc.Tuning.BlockProcessingFlushPeriod < 1 {
		rc.Tuning.BlockProcessingFlushPeriod = 3
	}
	if rc.Database != nil && rc.Database.CacheSize < 1 {
		log.Warn("Database cache size below limit", "old value", rc.Database.CacheSize, "new value", 10)
		rc.Database.CacheSize = 10
	}
	if rc.Connection.MaxReconnectTries > 0 && rc.Connection.ReconnectInterval < 1 {
		log.Warn("Quorum client reconnect interval below limit", "old value", rc.Connection.ReconnectInterval, "new value", 5)
		rc.Connection.ReconnectInterval = 5
	}
}

func ReadConfig(configFile string) (ReportingConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return ReportingConfig{}, err
	}
	defer f.Close()
	var input ReportingConfig
	if err = toml.NewDecoder(f).Decode(&input); err != nil {
		return ReportingConfig{}, err
	}

	input.SetDefaults()
	return input, nil
}
