package sign

import (
	"fmt"
	"testing"
)

func TestGenerateRsa(t *testing.T) {
	generator, _ := NewRsaGenerator(2048)
	privateKeyBase64, _ := generator.GenerateBase64KeyOfPrivateKey()
	fmt.Println("Private key base64: ", privateKeyBase64)
	pemPrivateKey, _ := generator.GeneratePemPrivateKey()
	fmt.Println("Private key pem: ", pemPrivateKey)
	pemPrivatePKCS1Key, _ := generator.GeneratePemPrivatePKCS1Key()
	fmt.Println("Private PKCS1 key pem: ", pemPrivatePKCS1Key)
	pemPublicKey, _ := generator.GeneratePemPublicKey()
	fmt.Println("Public key pem: ", pemPublicKey)
	pemPublicPKIXKey, _ := generator.GeneratePemPublicPKIXKey()
	fmt.Println("Public PKIX key pem: ", pemPublicPKIXKey)

}
