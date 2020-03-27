package main

import (
	"flag"
	"log"
	"quorumengineering/quorum-report/core"
	"quorumengineering/quorum-report/types"
)

func main() {
	// expects one input which the config file
	// read the config file path
	var configFile string
	flag.StringVar(&configFile, "config", "", "config file")
	flag.Parse()

	if configFile == "" {
		log.Panic("Config file path not given. Cannot start the service.")
	}

	// read the given config file
	config := types.ReadConfig(configFile)

	// start the back end with given config
	backend, err := core.New(config)
	if err != nil {
		log.Println("Exiting...")
		log.Panicf("error: %v", err)
		return
	}

	backend.Start()
}
