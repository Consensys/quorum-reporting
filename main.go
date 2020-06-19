package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"quorumengineering/quorum-report/core"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
	"quorumengineering/quorum-report/ui"
)

func main() {
	err := run()
	log.Info("Exiting")
	if err != nil {
		log.Error("error occurred in startup", "err", err.Error())
		os.Exit(1)
	}
}

func run() error {
	// Set up logging with given verbosity
	var verbosity int
	flag.IntVar(&verbosity, "verbosity", log.InfoLevel, "logging verbosity")
	// Read config file path
	var configFile string
	flag.StringVar(&configFile, "config", "config.toml", "config file")
	flag.Parse()

	logrus.SetLevel(logrus.Level(verbosity + 2))
	if configFile == "" {
		return errors.New("config file path not given")
	}

	log.Info("Config file found", "filename", configFile)

	// read the given config file
	config, err := types.ReadConfig(configFile)
	if err != nil {
		log.Error("Unable to read configuration", "err", err)
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
