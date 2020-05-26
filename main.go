package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"

	"quorumengineering/quorum-report/core"
	"quorumengineering/quorum-report/types"
)

func main() {
	// expects one input which the config file
	// read the config file path
	var configFile string
	flag.StringVar(&configFile, "config", "config.toml", "config file")
	flag.Parse()

	if configFile == "" {
		log.Panic("Config file path not given. Cannot start the service.")
	}

	// read the given config file
	config, err := types.ReadConfig(configFile)
	if err != nil {
		log.Panicf("unable to read configuration from the config file: %v", err)
	}

	// start the back end with given config
	backend, err := core.New(config)
	if err != nil {
		log.Println("Exiting...")
		log.Panicf("error: %v", err)
		return
	}

	backend.Start()
	defer backend.Stop()

	if config.Server.UIPort > 0 {
		// start a light weighted sample sample ui
		router := gin.Default()
		router.Use(static.Serve("/", static.LocalFile("./ui", true)))
		err := router.Run(":" + strconv.Itoa(config.Server.UIPort))
		if err != nil {
			log.Panicf("unable to start UI: %v", err)
		}
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)
	<-sigc
	log.Println("Got interrupted, shutting down...")
}
