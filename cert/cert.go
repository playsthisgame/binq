package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func Setup(path string) (string, string, error) {
	// create .cert if not exists
	_ = os.MkdirAll(path, os.ModePerm)

	// set cert and key paths
	certPath := filepath.Join(path, "cert.pem")
	keyPath := filepath.Join(path, "key.pem")

	// generate cert if missing
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return generateSelfSignedCert(certPath, keyPath)
	} else {
		return certPath, keyPath, nil
	}
}

func generateSelfSignedCert(certPath, keyPath string) (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	serialNumber, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Vibe Drive"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // valid 1 year

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&priv.PublicKey,
		priv,
	)
	if err != nil {
		return "", "", err
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return "", "", err
	}
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return "", "", err
	}
	defer keyOut.Close()

	pem.Encode(
		keyOut,
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)},
	)

	return certOut.Name(), keyOut.Name(), nil
}
