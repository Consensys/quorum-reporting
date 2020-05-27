package types

import (
	"errors"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/naoina/toml"
)

type ElasticsearchConfig struct {
	Addresses []string `toml:"urls,omitempty"`
	CloudID   string   `toml:"cloudid"`

	Username string `toml:"username"`
	Password string `toml:"password"`
	APIKey   string `toml:"apikey"`

	// PEM-encoded certificate authorities.
	// When set, an empty certificate pool will be created, and the certificates will be appended to it.
	// The option is only valid when the transport is not specified, or when it's http.Transport.
	//CACert []byte

	//RetryOnStatus        []int // List of status codes for retry. Default: 502, 503, 504.
	//DisableRetry         bool  // Default: false.
	//EnableRetryOnTimeout bool  // Default: false.
	//MaxRetries           int   // Default: 3.

	//DiscoverNodesOnStart  bool          // Discover nodes when initializing the client. Default: false.
	//DiscoverNodesInterval time.Duration // Discover nodes periodically. Default: disabled.
}

type DatabaseConfig struct {
	Elasticsearch *ElasticsearchConfig `toml:"elasticsearch,omitempty"`
	CacheSize     int                  `toml:"cacheSize,omitempty"`
}

type TuningConfig struct {
	BlockProcessingQueueSize int `toml:"blockProcessingQueueSize"`
}

type ReportingConfig struct {
	Title     string
	Addresses []common.Address `toml:"addresses,omitempty"`
	Database  *DatabaseConfig  `toml:"database,omitempty"`
	Server    struct {
		RPCAddr     string   `toml:"rpcAddr"`
		RPCCorsList []string `toml:"rpcCorsList,omitempty"`
		RPCVHosts   []string `toml:"rpcvHosts,omitempty"`
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
		rc.Tuning.BlockProcessingQueueSize = 100
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

	// if AlwaysReconnect is set to true, check if ReconnectInterval
	// and MaxReconnectTries are given or not. If not throw error
	if input.Connection.MaxReconnectTries > 0 && input.Connection.ReconnectInterval == 0 {
		return ReportingConfig{}, errors.New("ReconnectInterval should be greater than zero if MaxReconnectTries is set")
	}

	input.SetDefaults()
	return input, nil
}
