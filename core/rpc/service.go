package rpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
	"github.com/rs/cors"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

const (
	ReadTimeout  = 30 * time.Second
	WriteTimeout = 30 * time.Second
	IdleTimeout  = 120 * time.Second
)

type RPCService struct {
	cors        []string
	httpAddress string
	db          database.Database

	httpServer *http.Server
}

func NewRPCService(db database.Database, config types.ReportingConfig) *RPCService {
	return &RPCService{
		cors:        config.Server.RPCCorsList,
		httpAddress: config.Server.RPCAddr,
		db:          db,
	}
}

func (r *RPCService) Start() error {
	log.Info("Starting JSON-RPC server")

	jsonrpcServer := rpc.NewServer()
	jsonrpcServer.RegisterCodec(json.NewCodec(), "application/json")
	if err := jsonrpcServer.RegisterService(NewRPCAPIs(r.db), "reporting"); err != nil {
		return err
	}

	serverWithCors := cors.New(cors.Options{AllowedOrigins: r.cors}).Handler(jsonrpcServer)
	r.httpServer = &http.Server{
		Addr:    r.httpAddress,
		Handler: serverWithCors,

		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  IdleTimeout,
	}

	go func() {
		if err := r.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Unable to start JSON-RPC server", "err", err)
		}
	}()

	log.Info("JSON-RPC HTTP endpoint opened", "url", fmt.Sprintf("http://%s", r.httpServer.Addr))
	return nil
}

func (r *RPCService) Stop() {
	log.Info("Stopping JSON-RPC server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.httpServer.Shutdown(ctx); err != nil {
		log.Error("JSON-RPC server shutdown failed", "err", err)
	}

	log.Info("RPC HTTP endpoint closed", "url", fmt.Sprintf("http://%s", r.httpServer.Addr))
	log.Info("RPC service stopped")
}
