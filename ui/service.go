package ui

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
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
	// start a light weighted sample sample ui
	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./ui", true)))

	handler.srv = &http.Server{
		Addr:    ":" + strconv.Itoa(handler.port),
		Handler: router,
	}

	go func() {
		if err := handler.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("unable to start UI: %v", err)
		}
	}()
}

func (handler *UIHandler) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := handler.srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	return nil
}
