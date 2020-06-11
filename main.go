package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"quorumengineering/quorum-report/core"
	"quorumengineering/quorum-report/types"
	"quorumengineering/quorum-report/ui"
)

func main() {
	err := run()
	log.Println("Exiting...")
	if err != nil {
		_, _ = fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// expects one input which the config file
	// read the config file path
	var configFile string
	flag.StringVar(&configFile, "config", "config.toml", "config file")
	flag.Parse()

	if configFile == "" {
		return errors.New("config file path not given")
	}

	// read the given config file
	config, err := types.ReadConfig(configFile)
	if err != nil {
		return fmt.Errorf("unable to read configuration from the config file: %v", err)
	}

	// start the back end with given config
	backend, err := core.New(config)
	if err != nil {
		return fmt.Errorf("initialize backend error: %v", err)
	}

	err = backend.Start()
	defer backend.Stop()
	if err != nil {
		return err
	}

	if config.Server.UIPort > 0 {
		// start a light weighted sample sample ui
		uiHandler := ui.NewUIHandler(config.Server.UIPort)
		uiHandler.Start()
		defer uiHandler.Stop()
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)
	<-sigc
	log.Println("Got interrupted, shutting down...")
	return nil
}
