package types

import (
	"errors"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/naoina/toml"

	"quorumengineering/quorum-report/log"
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
	From         uint64         `toml:"from,omitempty""`
}

type TemplateConfig struct {
	TemplateName  string `toml:"templateName,omitempty"`
	ABI           string `toml:"abi,omitempty"`
	StorageLayout string `toml:"storageLayout,omitempty"`
}

type RuleConfig struct {
	Scope        string         `toml:"scope,omitempty"`
	Deployer     common.Address `toml:"deployer,omitempty"`
	TemplateName string         `toml:"templateName,omitempty"`
	EIP165       string         `toml:"eip165,omitempty"`
}

type ReportingConfig struct {
	Title     string
	Addresses []*AddressConfig  `toml:"addresses,omitempty"`
	Templates []*TemplateConfig `toml:"templates,omitempty"`
	Rules     []*RuleConfig     `toml:"rules,omitempty"`
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
	// validate config rules
	if err = input.Validate(); err != nil {
		return ReportingConfig{}, err
	}

	input.SetDefaults()
	return input, nil
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

func (rc *ReportingConfig) Validate() error {
	for _, template := range rc.Templates {
		if template.TemplateName == "" {
			return errors.New(fmt.Sprintf("empty template name: %v", template))
		}
		if template.ABI == "" {
			return errors.New(fmt.Sprintf("empty template ABI: %v", template))
		}
	}
	for _, rule := range rc.Rules {
		if rule.Scope != AllScope && rule.Scope != InternalScope && rule.Scope != ExternalScope {
			return errors.New(fmt.Sprintf("invalid rule scope: %v", rule))
		}
		if rule.TemplateName == "" {
			return errors.New(fmt.Sprintf("invalid rule template name: %v", rule))
		}
	}
	return nil
}
