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

type ReportInputStruct struct {
	Title     string
	Addresses []common.Address `toml:"addresses,omitempty"`
	Database  struct {
		Elasticsearch *ElasticsearchConfig `toml:"elasticsearch,omitempty"`
	}
	Server struct {
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
}

func ReadConfig(configFile string) (ReportInputStruct, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return ReportInputStruct{}, err
	}
	defer f.Close()
	var input ReportInputStruct
	if err := toml.NewDecoder(f).Decode(&input); err != nil {
		return ReportInputStruct{}, err
	}

	// if AlwaysReconnect is set to true, check if ReconnectInterval
	// and MaxReconnectTries are given or not. If not throw error
	if input.Connection.MaxReconnectTries > 0 && input.Connection.ReconnectInterval == 0 {
		return ReportInputStruct{}, errors.New("reconnection details not set properly in the config file")
	}
	return input, nil
}
