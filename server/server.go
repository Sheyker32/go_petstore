package server

import (
	"context"
	"fmt"

	"log"
	"net/http"
	"time"

	_ "swagger_petstore/cmd/docs"
)

type Server struct {
	HttpServer *http.Server
}

func NewServer(r http.Handler) *Server {
	s := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	server := &Server{
		HttpServer: s,
	}
	return server
}

func (s *Server) Serve(ctx context.Context) error {
	var err error

	chErr := make(chan error)
	go func() {
		fmt.Println("Starting server...")
		if err := s.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	select {
	case <-chErr:
		return err
	case <-ctx.Done():
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = s.HttpServer.Shutdown(ctxShutdown)

	return err
}
