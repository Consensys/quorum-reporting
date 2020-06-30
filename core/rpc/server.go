package rpc

import (
	"github.com/rs/cors"
	"net/http"
	"quorumengineering/quorum-report/types"
	"time"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"

	"quorumengineering/quorum-report/database"
)

const (
	ReadTimeout  = 30 * time.Second
	WriteTimeout = 30 * time.Second
	IdleTimeout  = 120 * time.Second
)

func MakeServer(db database.Database, config types.ReportingConfig) {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(NewRPCAPIs(db), "reporting")

	handler := cors.New(cors.Options{AllowedOrigins: config.Server.RPCCorsList}).Handler(s)

	srv := &http.Server{
		Addr:    config.Server.RPCAddr,
		Handler: handler,

		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  IdleTimeout,
	}

	go srv.ListenAndServe()
}
