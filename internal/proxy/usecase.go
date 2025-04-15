package proxy

import (
	"crypto/tls"
)

type Usecase interface {
	GetCertificate(domain string) (tls.Certificate, error)
}
