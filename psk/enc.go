package psk

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"io/ioutil"
	"strings"
)

var (
	// Key bytestrings are encoded as base64. The "/" character is used as a
	// delimeter somewhere else so the URL encoding variant is used.
	Encoding = base64.URLEncoding

	// The default curve used for unmarshalling.
	Curve = elliptic.P256()
)

// Reads a P-256 private key in ASN.1 format encoded with Encoding (so, almost
// PEM).
func ReadPrivateKeyFile(filename string) (*ecdsa.PrivateKey, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return PrivateFromString(strings.TrimSpace(string(b)))
}

func PrivateToString(sk *ecdsa.PrivateKey) (string, error) {
	der, err := x509.MarshalECPrivateKey(sk)
	if err != nil {
		return "", err
	}
	return Encoding.EncodeToString(der), nil
}

func PrivateFromString(s string) (*ecdsa.PrivateKey, error) {
	der, err := Encoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(der)
}

// Reads a file containing the ANSI curve points of a P-256 curve encoded with
// Encoding.
func ReadPublicKeyFile(filename string) (*ecdsa.PublicKey, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return PublicFromString(strings.TrimSpace(string(b)))
}

func PublicToString(pk *ecdsa.PublicKey) string {
	return Encoding.EncodeToString(elliptic.Marshal(pk.Curve, pk.X, pk.Y))
}

// Expects P-256 curve.
func PublicFromString(s string) (*ecdsa.PublicKey, error) {
	ansi, err := Encoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.Unmarshal(Curve, ansi)
	return &ecdsa.PublicKey{
		Curve: Curve,
		X:     x,
		Y:     y,
	}, nil
}
