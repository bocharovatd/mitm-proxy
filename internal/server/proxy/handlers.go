package proxy

import (
	proxyHandlers "github.com/bocharovatd/mitm-proxy/internal/proxy/delivery/proxy"
	proxyRepository "github.com/bocharovatd/mitm-proxy/internal/proxy/repository"
	proxyUsecase "github.com/bocharovatd/mitm-proxy/internal/proxy/usecase"
	requestRepository "github.com/bocharovatd/mitm-proxy/internal/request/repository"
	requestUsecase "github.com/bocharovatd/mitm-proxy/internal/request/usecase"
)

func (p *Proxy) MapHandlers() {
	proxyRepo := proxyRepository.NewProxyRepository(p.mongoClient)
	proxyUC := proxyUsecase.NewProxyUsecase(proxyRepo)
	requestRepo := requestRepository.NewRequestRepository(p.mongoClient)
	requestUC := requestUsecase.NewRequestUsecase(requestRepo)
	proxyH := proxyHandlers.NewProxyHandlers(proxyUC, requestUC)
	p.handlers = proxyH
}
