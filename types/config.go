package types

import (
	"errors"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/naoina/toml"
)

type ReportInputStruct struct {
	Title     string
	Addresses []common.Address `toml:"addresses,omitempty"`
	Database  struct {
		// TODO: placeholder
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
