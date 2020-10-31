package sign

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/gob"
	"encoding/pem"
)

// RsaGenerator generate rsa
type RsaGenerator struct {
	privateKey rsa.PrivateKey
	publicKey  rsa.PublicKey
	bitSize    int
}

// NewRsaGenerator new generator
func NewRsaGenerator(bitSize int) (*RsaGenerator, error) {
	reader := rand.Reader
	privateKey, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return nil, err
	}
	generator := &RsaGenerator{}
	publicKey := privateKey.PublicKey
	generator.publicKey = publicKey
	generator.privateKey = *privateKey
	generator.bitSize = bitSize
	return generator, nil
}

// NewRsa2048Generator new
func NewRsa2048Generator() (*RsaGenerator, error) {
	return NewRsaGenerator(2048)
}

// NewRsa3072Generator new
func NewRsa3072Generator() (*RsaGenerator, error) {
	return NewRsaGenerator(3072)
}

// NewRsa4096Generator new
func NewRsa4096Generator() (*RsaGenerator, error) {
	return NewRsaGenerator(4096)
}

// GenerateKeyOfPrivateKey generate private key
func (g *RsaGenerator) GenerateKeyOfPrivateKey() ([]byte, error) {
	return saveGobKey(g.privateKey)
}

// GenerateBase64KeyOfPrivateKey generate base64 private key
func (g *RsaGenerator) GenerateBase64KeyOfPrivateKey() (string, error) {
	key, e := g.GenerateKeyOfPrivateKey()
	if e != nil {
		return "", e
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// GeneratePemPrivateKey generate pem private key
func (g *RsaGenerator) GeneratePemPrivateKey() (string, error) {
	return savePEMKey(&g.privateKey)
}

// GeneratePemPrivatePKCS1Key pkcs1 key
func (g *RsaGenerator) GeneratePemPrivatePKCS1Key() (string, error) {
	return savePEMPrivatePKCS1Key(&g.privateKey)
}

// GeneratePemPublicKey pem public key
func (g *RsaGenerator) GeneratePemPublicKey() (string, error) {
	return savePublicPEMKey(g.publicKey)
}

// GeneratePemPublicPKIXKey pkix key
func (g *RsaGenerator) GeneratePemPublicPKIXKey() (string, error) {
	return savePublicPKIXKey(g.publicKey)
}

func saveGobKey(key interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}

	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(key); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func savePEMKey(key *rsa.PrivateKey) (string, error) {
	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return PemToString(privateKey)
}

func savePEMPrivatePKCS1Key(privateKey *rsa.PrivateKey) (string, error) {
	asn1Bytes := x509.MarshalPKCS1PrivateKey(privateKey)
	var pemkey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: asn1Bytes,
	}
	return PemToString(pemkey)
}

func savePublicPEMKey(pubkey rsa.PublicKey) (string, error) {
	asn1Bytes, err := asn1.Marshal(pubkey)
	if err != nil {
		return "", err
	}
	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	return PemToString(pemkey)
}

func savePublicPKIXKey(pubkey rsa.PublicKey) (string, error) {
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&pubkey)
	if err != nil {
		return "", err
	}
	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}
	return PemToString(pemkey)

}

// PemToBytes pem to byte array
func PemToBytes(b *pem.Block) ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := pem.Encode(buffer, b)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// PemToString pem to string
func PemToString(b *pem.Block) (string, error) {
	buffer := &bytes.Buffer{}
	err := pem.Encode(buffer, b)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
