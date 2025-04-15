package main

import (
	"context"
	"log"
	"sync"

	"github.com/bocharovatd/mitm-proxy/internal/pkg/db/mongo"
	httpServer "github.com/bocharovatd/mitm-proxy/internal/server/http"
	proxyServer "github.com/bocharovatd/mitm-proxy/internal/server/proxy"
)

func main() {
	mongoClient, err := mongo.New()
	if err != nil {
		log.Fatalf("error creating mongo client", err)
	}
	log.Println("Mongo client created")
	defer mongoClient.Disconnect(context.TODO())

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		proxyServer := proxyServer.New(mongoClient)
		err := proxyServer.Run()
		if err != nil {
			log.Fatalf("Error starting MITM proxy:", err)
		}
		log.Println("MITM Proxy stopped:", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer := httpServer.New(mongoClient)
		err = httpServer.Run()
		if err != nil {
			log.Fatalf("failed ro run API web server: ", err)
		}
		log.Println("API web server stopped:", err)
	}()

	wg.Wait()
}
