package proxy

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"time"

	"github.com/bocharovatd/mitm-proxy/internal/proxy"
)

type ProxyUsecase struct {
	proxyRepository proxy.Repository
}

func NewProxyUsecase(proxyRepo proxy.Repository) proxy.Usecase {
	return &ProxyUsecase{
		proxyRepository: proxyRepo,
	}
}

func (usecase *ProxyUsecase) GetCertificate(domain string) (tls.Certificate, error) {
	cert, err := usecase.proxyRepository.GetCertificateByDomain(domain)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to check certificate: %w", err)
	}

	if cert != nil {
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil && x509Cert.NotAfter.After(time.Now().Add(24*time.Hour)) {
			return *cert, nil
		}
	}

	newCert, err := usecase.generateCertificate(domain)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate certificate: %w", err)
	}

	if err := usecase.proxyRepository.SaveCertificate(domain, newCert); err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to save certificate: %w", err)
	}

	return newCert, nil
}

func (usecase *ProxyUsecase) generateCertificate(domain string) (tls.Certificate, error) {
	scriptPath := "internal/scripts/gen_cert.sh"

	err := os.Chmod(scriptPath, 0755)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to make script executable: %w", err)
	}

	serial := big.NewInt(time.Now().UnixNano())
	cmd := exec.Command(scriptPath, domain, fmt.Sprintf("%d", serial))

	var certOut bytes.Buffer
	cmd.Stdout = &certOut

	if err := cmd.Run(); err != nil {
		return tls.Certificate{}, fmt.Errorf("script execution failed: %w", err)
	}

	certPEM := certOut.Bytes()

	keyPEM, err := os.ReadFile("certs/cert.key")
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to read key file: %w", err)
	}

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create key pair: %w", err)
	}

	return tlsCert, nil
}
