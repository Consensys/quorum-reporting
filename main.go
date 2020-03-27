package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)
	<-sigc
	log.Println("Got interrupted, shutting down...")

	backend.Stop()
}
