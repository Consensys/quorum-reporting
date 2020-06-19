package ui

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"

	"quorumengineering/quorum-report/log"
)

type UIHandler struct {
	port int

	srv *http.Server

	mu sync.Mutex
}

func NewUIHandler(port int) *UIHandler {
	return &UIHandler{port: port}
}

func (handler *UIHandler) Start() {
	log.Info("Start UI", "port number", handler.port)

	// start a light weighted sample sample ui
	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./ui/build", true)))

	handler.srv = &http.Server{
		Addr:    ":" + strconv.Itoa(handler.port),
		Handler: router,
	}

	go func() {
		if err := handler.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Unable to start UI", "err", err)
		}
	}()
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
