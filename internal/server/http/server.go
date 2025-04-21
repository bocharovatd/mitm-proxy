package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gorilla/mux"
)

const (
	ctxTimeout = 5
)

type Server struct {
	MUX         *mux.Router
	mongoClient *mongo.Client
}

func New(mongoClient *mongo.Client) *Server {
	return &Server{MUX: mux.NewRouter(), mongoClient: mongoClient}
}

func (s *Server) Run() error {
	s.MapHandlers()

	server := &http.Server{
		Addr:         ":8000",
		Handler:      s.MUX,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Starting API web server on :8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Error ListenAndServe in API web server: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), ctxTimeout*time.Second)
	defer shutdown()

	log.Println("API web server graceful shutdown")
	return server.Shutdown(ctx)
}
