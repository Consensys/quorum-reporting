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
	// Get Licenses
	var showLicenses bool
	flag.BoolVar(&showLicenses, "licenses", false, "show licenses")
	flag.Parse()

	if showLicenses {
		fmt.Println("Copyright 2020 JP Morgan Chase Company")
		fmt.Println()
		fmt.Println("Quorum Reporting is licensed under the Apache License, Version 2.0 (the \"License\").")
		fmt.Println("You may obtain a copy of the License at https://github.com/QuorumEngineering/quorum-test/LICENSE.")
		fmt.Println()
		fmt.Println("Quorum Reporting is made possible by Quorum Reporting open source project and other open source libraries including:")
		fmt.Println("github.com/bluele/gcache                check license at: https://github.com/bluele/gcache/blob/master/LICENSE")
		fmt.Println("github.com/elastic/go-elasticsearch/v7  check license at: https://github.com/elastic/go-elasticsearch/blob/master/LICENSE")
		fmt.Println("github.com/gin-gonic/contrib            check license at: https://github.com/gin-gonic/contrib/blob/master/LICENSE")
		fmt.Println("github.com/gin-gonic/gin                check license at: https://github.com/gin-gonic/gin/blob/master/LICENSE")
		fmt.Println("github.com/golang/mock                  check license at: https://github.com/golang/mock/blob/master/LICENSE")
		fmt.Println("github.com/gorilla/rpc                  check license at: https://github.com/gorilla/rpc/blob/master/LICENSE")
		fmt.Println("github.com/gorilla/websocket            check license at: https://github.com/gorilla/websocket/blob/master/LICENSE")
		fmt.Println("github.com/machinebox/graphql           check license at: https://github.com/machinebox/graphql/blob/master/LICENSE")
		fmt.Println("github.com/matryer/is                   check license at: https://github.com/matryer/is/blob/master/LICENSE")
		fmt.Println("github.com/naoina/go-stringutil         check license at: https://github.com/naoina/go-stringutil/blob/master/LICENSE")
		fmt.Println("github.com/naoina/toml                  check license at: https://github.com/naoina/toml/blob/master/LICENSE")
		fmt.Println("github.com/pkg/errors                   check license at: https://github.com/pkg/errors/blob/master/LICENSE")
		fmt.Println("github.com/rs/cors                      check license at: https://github.com/rs/cors/blob/master/LICENSE")
		fmt.Println("github.com/sirupsen/logrus              check license at: https://github.com/sirupsen/logrus/blob/master/LICENSE")
		fmt.Println("github.com/stretchr/testify             check license at: https://github.com/stretchr/testify/blob/master/LICENSE")
		fmt.Println("golang.org/x/crypto                     check license at: https://golang.org/LICENSE")
		os.Exit(0)
	}

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
	select {
	case <-sigc:
	case <-backend.GetBackendErrorChannel(): //Check for errors that will warrant an application shutdown
	}
	log.Info("Received interrupt signal, shutting down...")
	return nil
}
