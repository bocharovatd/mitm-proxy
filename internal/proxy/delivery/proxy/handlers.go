package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/bocharovatd/mitm-proxy/internal/proxy"
	"github.com/bocharovatd/mitm-proxy/internal/request"
	requestEntity "github.com/bocharovatd/mitm-proxy/internal/request/entity"
)

type ProxyHandlers struct {
	usecase        proxy.Usecase
	requestUsecase request.Usecase
}

func NewProxyHandlers(proxyUC proxy.Usecase, requestUC request.Usecase) proxy.Handlers {
	return &ProxyHandlers{
		usecase:        proxyUC,
		requestUsecase: requestUC,
	}
}

func (handlers *ProxyHandlers) HandleConnection(conn net.Conn) {
	defer conn.Close()

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}

	if request.Method == http.MethodConnect {
		handlers.HandleHTTPSConnection(conn, request)
	} else {
		handlers.HandleHTTPConnection(conn, request, nil)
	}
}

func (handlers *ProxyHandlers) HandleHTTPConnection(conn net.Conn, request *http.Request, tlsConfig *tls.Config) {
	clientIP := conn.RemoteAddr().String()
	httpReq := requestEntity.ParseHTTPRequest(request)

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

	startTime := time.Now()

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

	duration := time.Since(startTime)

	httpResp := requestEntity.ParseHTTPResponse(response, duration)

	if _, err := handlers.requestUsecase.Save(httpReq, httpResp, clientIP); err != nil {
		log.Printf("Failed to save request: %v", err)
	}

	err = response.Write(conn)
	if err != nil {
		log.Println("Error sending response to client:", err)
		return
	}
}

func (handlers *ProxyHandlers) HandleHTTPSConnection(conn net.Conn, request *http.Request) {
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

	domain := request.URL.Hostname()

	cert, err := handlers.usecase.GetCertificate(domain)
	if err != nil {
		log.Printf("Failed to get certificate for %s: %v", domain, err)
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
		handlers.HandleHTTPConnection(tlsConn, request, tlsConfig)
	}
}
