package main

import (
	"log"
	"github.com/bocharovatd/mitm-proxy/internal/proxy"
)

func main() {
	err := proxy.StartProxyServer()
	if (err != nil) {
		log.Println("Error starting MITM proxy:", err)
	}
	log.Println("MITM Proxy stopped:", err)
}