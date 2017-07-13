package cipher

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

type RSA struct {
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

func ParseRsaPrivateKeyFromPem(privPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ParseRsaPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseRsaPrivateKeyFromPem(data)
}

func LoadRSAFromFile(filename string) (*RSA, error) {
	privKey, err := ParseRsaPrivateKeyFromFile(filename)
	if err != nil {
		return nil, err
	}
	privKey.Precompute()
	return &RSA{
		privKey: privKey,
		pubKey:  &privKey.PublicKey,
	}, nil
}

func LoadRSAFromPem(pem []byte) (*RSA, error) {
	privKey, err := ParseRsaPrivateKeyFromPem(pem)
	if err != nil {
		return nil, err
	}
	privKey.Precompute()
	return &RSA{
		privKey: privKey,
		pubKey:  &privKey.PublicKey,
	}, nil
}

func (r *RSA) Encrypt(plain []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, r.pubKey, plain)
}

func (r *RSA) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, r.privKey, ciphertext)
}
