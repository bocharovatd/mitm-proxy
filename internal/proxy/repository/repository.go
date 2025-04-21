package proxy

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bocharovatd/mitm-proxy/internal/proxy"
)

type ProxyRepository struct {
	mongoCollection *mongo.Collection
}

func NewProxyRepository(mongoClient *mongo.Client) proxy.Repository {
	collection := mongoClient.Database("MongoBD").Collection("certificates")
	return &ProxyRepository{
		mongoCollection: collection,
	}
}

type CertificateDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Domain    string             `bson:"domain"`
	CertPEM   string             `bson:"cert_pem"`
	KeyPEM    string             `bson:"key_pem"`
	CreatedAt time.Time          `bson:"created_at"`
	ExpiresAt time.Time          `bson:"expires_at"`
}

func (r *ProxyRepository) SaveCertificate(domain string, cert tls.Certificate) error {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Certificate[0],
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(cert.PrivateKey.(*rsa.PrivateKey)),
	})

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	doc := CertificateDocument{
		Domain:    domain,
		CertPEM:   string(certPEM),
		KeyPEM:    string(keyPEM),
		CreatedAt: time.Now(),
		ExpiresAt: x509Cert.NotAfter,
	}

	_, err = r.mongoCollection.InsertOne(context.Background(), doc)
	if err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	return nil
}

func (r *ProxyRepository) GetCertificateByDomain(domain string) (*tls.Certificate, error) {
	var doc CertificateDocument
	filter := bson.M{"domain": domain}

	err := r.mongoCollection.FindOne(context.Background(), filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	certBlock, _ := pem.Decode([]byte(doc.CertPEM))
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	keyBlock, _ := pem.Decode([]byte(doc.KeyPEM))
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	cert := tls.Certificate{
		Certificate: [][]byte{certBlock.Bytes},
		PrivateKey:  privateKey,
	}

	return &cert, nil
}
