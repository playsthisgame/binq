package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Setup generates a single certificate used for both local and LAN access.
func Setup(path string) error {
	_ = os.MkdirAll(path, os.ModePerm)

	certPath := filepath.Join(path, "vibedrive.pem")
	keyPath := filepath.Join(path, "vibedrive-key.pem")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		if err := generateSelfSignedCert(certPath, keyPath); err != nil {
			return fmt.Errorf("failed to generate certificate: %w", err)
		}
	}

	return nil
}

// generateSelfSignedCert creates a TLS certificate valid for localhost and LAN IPs.
func generateSelfSignedCert(certPath, keyPath string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serialNumber, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "VibeDrive",
			Organization: []string{"Vibe Drive"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // valid for 1 year

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames: []string{"localhost"},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
	}

	if localIP := getLocalIP(); localIP != nil {
		template.IPAddresses = append(template.IPAddresses, localIP)
		fmt.Println("📡 Detected LAN IP for cert:", localIP.String())
	} else {
		fmt.Println("⚠️ Warning: could not detect local IP — cert will only be valid for localhost")
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
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return nil
}

// getLocalIP finds the first non-loopback IPv4 address on the host.
func getLocalIP() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok &&
			!ipNet.IP.IsLoopback() &&
			ipNet.IP.To4() != nil {
			return ipNet.IP
		}
	}
	return nil
}
