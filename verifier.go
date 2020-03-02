package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
)

const rsaPrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBAKkOfVaHMu2s+JWpqECdwp7di1MBOC0tNQH1lYi0AL6qBHZmU6BJ
MBFe7JU/9ctanTW+3X/a345XalZd1u0Z9R0CAwEAAQJBAIyf6+y1K8z/DAzagoW1
dTXXDdTu977EkwpdMZT0PoZZ5qCFQMuNrFLbxNyPIdkpzfm3jpYAP/dyDb7DLHEP
X0ECIQDQbPwNFkyBuCOYvy2G3iTUgyZrK6LSV7xAwaBQYFkpDQIhAM+lCPNMpqWj
x5QE3M7rgJ0F1pPNx3r2woJJ2YDNBdhRAiAZztTq/e7dRSLLQCjwAUPIOLEiJhYU
O57o2dDzAusnZQIgHiwZ/9iMgpco4f5O45Ze6vI1Oub07I48t1fpzgh8p/ECIQCt
jh3wKn0HGIjep4ZKpYVDEH8+dlZJ0J7fkv4YH3D+XA==
-----END RSA PRIVATE KEY-----
`

const rsaPublicKey = `
-----BEGIN RSA PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKkOfVaHMu2s+JWpqECdwp7di1MBOC0t
NQH1lYi0AL6qBHZmU6BJMBFe7JU/9ctanTW+3X/a345XalZd1u0Z9R0CAwEAAQ==
-----END RSA PUBLIC KEY-----
`

// ParseRsaPrivateKeyFromPemStr allows to deserialize a RSA private key from a string
func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

// ParseRsaPublicKeyFromPemStr allows to deserialize a RSA public key from a string
func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("Key type is not RSA")
}

// VerifyRSASignature allows to verfiy files RSA signature
func VerifyRSASignature(targetFilename, signatureFilename string) (err error) {
	pubKey, err := ParseRsaPublicKeyFromPemStr(rsaPublicKey)
	if err != nil {
		return
	}
	data, _ := ioutil.ReadFile(targetFilename)
	if err != nil {
		return
	}
	digest := sha256.Sum256(data)
	signRead, _ := ioutil.ReadFile(signatureFilename)
	if err != nil {
		return
	}
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest[:], signRead)
	if err != nil {
		log.Fatalf("Verification failed: %s", err)
		return
	}

	return nil
}

// SignRSA allows to sign files with RSA
func SignRSA(targetFilename, outFilename string) (err error) {
	privKey, err := ParseRsaPrivateKeyFromPemStr(rsaPrivateKey)
	if err != nil {
		return
	}
	data, err := ioutil.ReadFile(targetFilename)
	if err != nil {
		return
	}

	digest := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, digest[:])
	if err != nil {
		return
	}

	err = ioutil.WriteFile(outFilename, signature, 0664)
	if err != nil {
		return
	}

	return nil
}

// GenerateKeyPair can be used for testing or exporting keys to file
func GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privkey, &privkey.PublicKey, nil
}
