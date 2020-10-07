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

type RsaGenerator struct {
	privateKey rsa.PrivateKey
	publicKey  rsa.PublicKey
	bitSize    int
}

func NewRsaGenerator(bitSize int) (*RsaGenerator, error) {
	reader := rand.Reader
	if privateKey, err := rsa.GenerateKey(reader, bitSize); err != nil {
		return nil, err
	} else {
		generator := &RsaGenerator{}
		publicKey := privateKey.PublicKey
		generator.publicKey = publicKey
		generator.privateKey = *privateKey
		generator.bitSize = bitSize
		return generator, nil
	}
}
func NewRsa2048Generator() (*RsaGenerator, error) {
	return NewRsaGenerator(2048)
}

func NewRsa3072Generator() (*RsaGenerator, error) {
	return NewRsaGenerator(3072)
}

func NewRsa4096Generator() (*RsaGenerator, error) {
	return NewRsaGenerator(4096)
}

func (g *RsaGenerator) GenerateKeyOfPrivateKey() ([]byte, error) {
	return saveGobKey(g.privateKey)
}

func (g *RsaGenerator) GenerateBase64KeyOfPrivateKey() (string, error) {
	if key, e := g.GenerateKeyOfPrivateKey(); e != nil {
		return "", e
	} else {
		return base64.StdEncoding.EncodeToString(key), nil
	}
}

func (g *RsaGenerator) GeneratePemPrivateKey() (string, error) {
	return savePEMKey(&g.privateKey)
}

func (g *RsaGenerator) GeneratePemPrivatePKCS1Key() (string, error) {
	return savePEMPrivatePKCS1Key(&g.privateKey)
}

func (g *RsaGenerator) GeneratePemPublicKey() (string, error) {
	return savePublicPEMKey(g.publicKey)
}

func (g *RsaGenerator) GeneratePemPublicPKIXKey() (string, error) {
	return savePublicPKIXKey(g.publicKey)
}

func saveGobKey(key interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}

	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(key); err != nil {
		return nil, err
	} else {
		return buffer.Bytes(), nil
	}
}

func savePEMKey(key *rsa.PrivateKey) (string, error) {
	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return pemToString(privateKey)
}

func savePEMPrivatePKCS1Key(privateKey *rsa.PrivateKey) (string, error) {
	asn1Bytes := x509.MarshalPKCS1PrivateKey(privateKey)
	var pemkey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: asn1Bytes,
	}
	return pemToString(pemkey)
}

func savePublicPEMKey(pubkey rsa.PublicKey) (string, error) {
	if asn1Bytes, err := asn1.Marshal(pubkey); err != nil {
		return "", err
	} else {
		var pemkey = &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: asn1Bytes,
		}

		return pemToString(pemkey)
	}
}

func savePublicPKIXKey(pubkey rsa.PublicKey) (string, error) {
	if asn1Bytes, err := x509.MarshalPKIXPublicKey(&pubkey); err != nil {
		return "", err
	} else {
		var pemkey = &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: asn1Bytes,
		}
		return pemToString(pemkey)
	}

}

func pemToBytes(b *pem.Block) ([]byte, error) {
	buffer := &bytes.Buffer{}
	if err := pem.Encode(buffer, b); err != nil {
		return nil, err
	} else {
		return buffer.Bytes(), nil
	}
}

func pemToString(b *pem.Block) (string, error) {
	buffer := &bytes.Buffer{}
	if err := pem.Encode(buffer, b); err != nil {
		return "", err
	} else {
		return string(buffer.Bytes()), nil
	}
}
