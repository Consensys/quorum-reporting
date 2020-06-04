package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"quorumengineering/quorum-report/core"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
	"quorumengineering/quorum-report/ui"
)

func main() {
	err := run()
	log.Info("Exiting")
	if err != nil {
		log.Error(err.Error())
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

	log.Info("config file found", "filename", configFile)

	// read the given config file
	config, err := types.ReadConfig(configFile)
	if err != nil {
		log.Error("unable to read configuration", "err", err)
		return errors.New("unable to read configuration")
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

	log.Debug("UI Port", "port number", config.Server.UIPort)
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
	log.Info("Received interrupt signal, shutting down...")
	return nil
}
