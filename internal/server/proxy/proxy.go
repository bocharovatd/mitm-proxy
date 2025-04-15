package proxy

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bocharovatd/mitm-proxy/internal/proxy"
)

type Proxy struct {
	handlers    proxy.Handlers
	mongoClient *mongo.Client
}

func New(mongoClient *mongo.Client) *Proxy {
	return &Proxy{mongoClient: mongoClient}
}

func (p *Proxy) Run() error {
	p.MapHandlers()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Println("MITM Proxy started on :8080")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		cancel()
		listener.Close()
	}()

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down MITM proxy...")
			wg.Wait()
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					continue
				}
				log.Println("Accept error:", err)
				continue
			}

			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				p.handlers.HandleConnection(c)
			}(conn)
		}
	}
}
