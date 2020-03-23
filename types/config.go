package types

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hpcloud/tail/util"
	"github.com/naoina/toml"
)

type ReportInputStruct struct {
	Title     string
	Reporting struct {
		WSUrl       string
		GraphQLUrl  string
		Addresses   []common.Address
		RPCAddr     string
		RPCCorsList []string
		RPCVHosts   []string
	}
}

func ReadConfig(configFile string) ReportInputStruct {
	f, err := os.Open(configFile)
	if err != nil {
		util.Fatal("unable to open the config file %v", err)
	}
	defer f.Close()
	var input ReportInputStruct
	if err := toml.NewDecoder(f).Decode(&input); err != nil {
		util.Fatal("unable to read the config file %v", err)
	}
	return input
}
