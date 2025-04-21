package http

import (
	"net/http"

	requestHandlers "github.com/bocharovatd/mitm-proxy/internal/request/delivery/http"
	requestRepository "github.com/bocharovatd/mitm-proxy/internal/request/repository"
	requestUsecase "github.com/bocharovatd/mitm-proxy/internal/request/usecase"
)

func (s *Server) MapHandlers() {
	requestRepo := requestRepository.NewRequestRepository(s.mongoClient)
	requestUC := requestUsecase.NewRequestUsecase(requestRepo)
	requestH := requestHandlers.NewRequestHandlers(requestUC)
	s.MUX.Handle("/requests", http.HandlerFunc(requestH.GetAll)).Methods("GET")
	s.MUX.Handle("/requests/{requestID:[0-9a-fA-F]{24}}", http.HandlerFunc(requestH.GetByID)).Methods("GET")
	s.MUX.Handle("/repeat/{requestID:[0-9a-fA-F]{24}}", http.HandlerFunc(requestH.RepeatByID)).Methods("GET")
	s.MUX.Handle("/scan/{requestID:[0-9a-fA-F]{24}}", http.HandlerFunc(requestH.ScanByID)).Methods("GET")
}
