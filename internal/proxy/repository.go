package proxy

import (
	"crypto/tls"
)

type Repository interface {
	SaveCertificate(domain string, cert tls.Certificate) error
	GetCertificateByDomain(domain string) (*tls.Certificate, error)
}
