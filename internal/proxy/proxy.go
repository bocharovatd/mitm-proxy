package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http/httputil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"math/big"
	"time"

	// "crypto/x509"
	"os/exec"
	"bytes"
	"io"
	// "encoding/pem"
)

func StartProxyServer() error  {
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
			log.Println("Shutting down...")
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
				HandleConnection(c)
			}(conn)
		}
	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}

	if request.Method == http.MethodConnect {
		HandleHTTPSConnection(conn, request)
	} else {
		HandleHTTPConnection(conn, request, nil)
	}
}

func HandleHTTPConnection(conn net.Conn, request *http.Request, tlsConfig *tls.Config) {
	var targetConn net.Conn 
	var err error
	if tlsConfig != nil {
		targetConn, err = tls.Dial("tcp", fmt.Sprintf("%s:%s", request.Host, "443"), tlsConfig)
	} else {
		targetConn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", request.URL.Hostname(), "80"))
	}

	if err != nil {
		log.Println("Error connecting to target:", err)
		return
	}

	defer targetConn.Close()

	request.Header.Del("Proxy-Connection")
	request.RequestURI = ""

	dump, err := httputil.DumpRequest(request, true)
    if err != nil {
        log.Println("Error dumping request:", err)
    } else {
        log.Printf("Target request:\n%s", dump)
    }

	err = request.Write(targetConn)
	if err != nil {
		log.Println("Error sending request to target:", err)
		return
	}

	response, err := http.ReadResponse(bufio.NewReader(targetConn), request)
	if err != nil {
		log.Println("Error reading response:", err)
		return
	}
	defer response.Body.Close()

	err = response.Write(conn)
	if err != nil {
		log.Println("Error sending response to client:", err)
		return
	}
}

func GenerateCertFromScript(domain string, serial *big.Int) (tls.Certificate, error) {
	cmd := exec.Command("internal/scripts/gen_cert.sh", domain, fmt.Sprintf("%d", serial))
	
	var certOut bytes.Buffer
	cmd.Stdout = &certOut
	
	if err := cmd.Run(); err != nil {
		return tls.Certificate{}, fmt.Errorf("script execution failed: %w", err)
	}

	certPEM := certOut.Bytes()

	keyPEM, err := os.ReadFile("certs/cert.key")
    if err != nil {
        return tls.Certificate{}, err
    }

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)

	return tlsCert, nil
}

func HandleHTTPSConnection(conn net.Conn, request *http.Request) {
	dump, err := httputil.DumpRequest(request, true)
    if err != nil {
        log.Println("Error dumping request:", err)
    } else {
        log.Printf("CONNECT request:\n%s", dump)
    }

	_, err = conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		log.Printf("Failed to send CONNECT response: %v", err)
		return
	}

	cert, err := GenerateCertFromScript(request.URL.Hostname(), big.NewInt(time.Now().UnixNano()))
	if err != nil {
		log.Printf("Failed to generate certificate: %v", err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConn := tls.Server(conn, tlsConfig)
	defer tlsConn.Close()

	reader := bufio.NewReader(tlsConn)

	for {
		request, err = http.ReadRequest(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading request: %v", err)
			}
			return
		}
		HandleHTTPConnection(tlsConn, request, tlsConfig)
	}
}