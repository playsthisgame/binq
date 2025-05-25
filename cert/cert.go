package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func Setup(path string) error {
	// create .cert if not exists
	_ = os.MkdirAll(path, os.ModePerm)

	// set cert and key paths
	remoteCertPath := filepath.Join(path, "remotehost.pem")
	remoteKeyPath := filepath.Join(path, "remotehost-key.pem")
	localCertPath := filepath.Join(path, "localhost.pem")
	localKeyPath := filepath.Join(path, "localhost-key.pem")

	// generate cert if missing
	if _, err := os.Stat(remoteCertPath); os.IsNotExist(err) {
		generateSelfSignedCert(remoteCertPath, remoteKeyPath, false)
	}
	if _, err := os.Stat(localCertPath); os.IsNotExist(err) {
		generateSelfSignedCert(localCertPath, localKeyPath, true)
	}

	return nil
}

func generateSelfSignedCert(certPath, keyPath string, local bool) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
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

	if local {
		template.Subject.CommonName = "localhost"
		template.DNSNames = []string{"localhost"}
		template.IPAddresses = []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		}
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&priv.PublicKey,
		priv,
	)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	pem.Encode(
		keyOut,
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)},
	)

	return nil
}
