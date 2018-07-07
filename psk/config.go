package psk

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

type Config struct {
	ServerName string
	Key        *ecdsa.PrivateKey
	PeerKey    *ecdsa.PublicKey
}

func (c Config) ClientTLS() (*tls.Config, error) {
	tc := c.tlsConfig()
	certs, err := c.genCerts(false)
	if err != nil {
		return nil, err
	}
	tc.Certificates = certs
	return tc, nil
}

func (c Config) ServerTLS() (*tls.Config, error) {
	tc := c.tlsConfig()
	certs, err := c.genCerts(false)
	if err != nil {
		return nil, err
	}
	tc.Certificates = certs
	return tc, nil
}

const keyUsage = x509.KeyUsageDigitalSignature |
	x509.KeyUsageKeyEncipherment |
	x509.KeyUsageKeyAgreement |
	x509.KeyUsageCertSign

func (c *Config) genCerts(server bool) ([]tls.Certificate, error) {
	var extUsage x509.ExtKeyUsage
	if server {
		extUsage = x509.ExtKeyUsageServerAuth
	} else {
		extUsage = x509.ExtKeyUsageClientAuth
	}

	sk := c.Key
	pk := &sk.PublicKey
	now := time.Now()
	cert := &x509.Certificate{
		// x509v3 with serial number 1.
		Version:      3,
		SerialNumber: big.NewInt(1),

		// Self-signed, so Issuer == Subject. We only care about the CommonName,
		// but that's only for debugging.
		Issuer:  pkix.Name{CommonName: c.ServerName},
		Subject: pkix.Name{CommonName: c.ServerName},

		// Give an hour in the past for some absurd clock drift, and 5 years of
		// validity. We check neither of these, but might in the future.
		NotBefore: now.Add(-time.Hour),
		NotAfter:  now.Add(5 * 365 * 24 * time.Hour),

		// Give appropriate usage flags.
		KeyUsage:    keyUsage,
		ExtKeyUsage: []x509.ExtKeyUsage{extUsage},

		// This lets the x509 library validate the self-signed cert against
		// itself.
		BasicConstraintsValid: true,
		IsCA: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, pk, sk)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{
		tls.Certificate{
			Certificate: [][]byte{certBytes},
			PrivateKey:  sk,
		},
	}, nil
}

func (pskConfig *Config) tlsConfig() *tls.Config {
	return &tls.Config{
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) != 1 {
				return fmt.Errorf("expected one cert; got %v", len(rawCerts))
			}
			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return err
			}
			if cert.PublicKeyAlgorithm != x509.ECDSA {
				return fmt.Errorf("expected ECDSA public key; got %q", cert.PublicKeyAlgorithm.String())
			}
			switch cert.SignatureAlgorithm {
			case x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512:
			default:
				return fmt.Errorf("expected ECDSA signature; got %q", cert.SignatureAlgorithm.String())
			}

			pk, ok := cert.PublicKey.(*ecdsa.PublicKey)
			if !ok {
				return fmt.Errorf("public key not an ecdsa.PublicKey; got %T", cert.PublicKey)
			}

			// Check the public key is expected.
			if pk.Curve != pskConfig.PeerKey.Curve {
				return fmt.Errorf("public key had different curve; got %+v", pskConfig.PeerKey.Curve)
			}
			if pk.X.Cmp(pskConfig.PeerKey.X) != 0 || pk.Y.Cmp(pskConfig.PeerKey.Y) != 0 {
				return fmt.Errorf("public key mismatch")
			}

			// Certificate will be self-signed.
			return cert.CheckSignatureFrom(cert)
		},
		ServerName: pskConfig.ServerName,

		ClientAuth:         tls.RequireAnyClientCert,
		InsecureSkipVerify: true, // we verify certificates based on keys
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
		MinVersion: tls.VersionTLS12,
	}
}
