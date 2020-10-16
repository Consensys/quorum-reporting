package ui

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"

	"quorumengineering/quorum-report/log"
	_ "quorumengineering/quorum-report/ui/statik" //allow the packages `init` function to run, registered the asset data
)

type UIHandler struct {
	port int

	srv *http.Server

	mu sync.Mutex
}

func NewUIHandler(port int) *UIHandler {
	return &UIHandler{port: port}
}

func (handler *UIHandler) Start() error {
	log.Info("Start UI", "port number", handler.port)

	statikFS, err := fs.New()
	if err != nil {
		return err
	}

	// start a light weighted sample sample ui
	router := gin.Default()
	router.StaticFS("/", statikFS)
	router.NoRoute(func(c *gin.Context) {
		// React-router uses fake urls to handle deep links and routing in the ui. If the user goes
		// directly to a non-root path it will 404. Instead, just silently serve them the root html
		c.Request.URL.Path = "/"
		router.HandleContext(c)
	})

	handler.srv = &http.Server{
		Addr:    ":" + strconv.Itoa(handler.port),
		Handler: router,
	}

	go func() {
		if err := handler.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Unable to start UI", "err", err)
		}
	}()
	return nil
}

func (handler *UIHandler) Stop() error {
	log.Info("Stopping UI server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := handler.srv.Shutdown(ctx); err != nil {
		log.Error("UI server shutdown failed", "err", err)
	}
	return nil
}
