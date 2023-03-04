package authorization

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

//go:generate go run github.com/golang/mock/mockgen -source=certificate.go -destination=mock/certificate_mock.go CertificateManager

// Certificate 存储在配置内的证书
type Certificate struct {
	SerialNumber string
	PublicKey    string
	PrivateKey   string
	ExpireTime   time.Time
}

// AuthInfo http头内的鉴权信息
type AuthInfo struct {
	Timestamp  string `json:"timestamp"`
	Nonce      string `json:"nonce"`
	Signature  string `json:"signature"`
	SerialNo   string `json:"serial_no"`
	MerchantID string `json:"merchant_id"`
}

// CertificateManager 证书管理
type CertificateManager interface {
	// GetMerchantCertificate 获取商户证书
	GetMerchantCertificate(ctx context.Context, merchantID string, serialNumber string) (*x509.Certificate, error)
	// GetPlatformCertificate 获取平台证书
	GetPlatformCertificate(ctx context.Context) (*x509.Certificate, error)
}

func GenerateCertificate() (*x509.Certificate, error) {
	// 生成RSA密钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// 准备证书请求信息
	subject := pkix.Name{
		CommonName:         "example.com",
		Organization:       []string{"Example Company"},
		OrganizationalUnit: []string{"IT"},
		Country:            []string{"US"},
		Province:           []string{"California"},
		Locality:           []string{"San Francisco"},
	}

	// 生成证书请求
	// csrTemplate := x509.CertificateRequest{
	// 	Subject:            subject,
	// 	SignatureAlgorithm: x509.SHA256WithRSA,
	// }
	// csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	// if err != nil {
	// 	panic(err)
	// }

	// 生成自签名证书
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(err)
	}
	certificateTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	certificateBytes, err := x509.CreateCertificate(
		rand.Reader, &certificateTemplate, &certificateTemplate, &privateKey.PublicKey, privateKey,
	)
	if err != nil {
		panic(err)
	}
	certificates, err := x509.ParseCertificates(certificateBytes)
	if err != nil {
		return nil, err
	}
	return certificates[0], nil

	// // 将密钥和证书写入文件
	// privateKeyFile, err := os.Create("private.key")
	// if err != nil {
	// 	panic(err)
	// }
	// defer privateKeyFile.Close()
	// err = pem.Encode(privateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	// if err != nil {
	// 	panic(err)
	// }
	//
	// certificateFile, err := os.Create("certificate.crt")
	// if err != nil {
	// 	panic(err)
	// }
	// defer certificateFile.Close()
	// err = pem.Encode(certificateFile, &pem.Block{Type: "CERTIFICATE", Bytes: certificateBytes})
	// if err != nil {
	// 	panic(err)
	// }
	//
	// csrFile, err := os.Create("certificate.csr")
	// if err != nil {
	// 	panic(err)
	// }
	// defer csrFile.Close()
	// err = pem.Encode(csrFile, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	// if err != nil {
	// 	panic(err)
	// }
}
