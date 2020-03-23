package main

import (
	"flag"
	"github.com/hpcloud/tail/util"
	"log"
	"quorumengineering/quorum-report/core"
)

func main() {
	// expects one input which the config file
	// read the config file path
	var configFile string
	flag.StringVar(&configFile, "config", "", "config file")
	flag.Parse()

	if configFile == "" {
		util.Fatal("config file path not given. cannot start the service")
	}

	// read the given config file
	config := core.ReadConfig(configFile)

	// start the back end with given config
	backend, err := core.New(config)
	if err != nil {
		log.Fatalf("exiting... err: %v.\n", err)
		return
	}

	backend.Start()
}
