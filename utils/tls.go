package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func loadPem(p string) ([]byte, error) {
	pemBytes, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM file")
	}
	return block.Bytes, nil
}

func loadPrivKeyFromPEM(p string) (*rsa.PrivateKey, error) {
	derBytes, err := loadPem(p)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PrivateKey(derBytes)
}

func loadCertFromPEM(p string) (*x509.Certificate, error) {
	derBytes, err := loadPem(p)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(derBytes)
}

func GenerateCert(outKey string, outCert string, caKeyPath string, caCertPath string, hosts []string) (err error) {
	//if caKeyPath == "" {
	//	logs.Get().Infof("generate CA certificate, %s, %s", outKey, outCert)
	//} else {
	//	logs.Get().Infof("generate certificate, %s, %s", outKey, outCert)
	//}

	for _, s := range []string{outKey, outCert} {
		if FileExist(s) {
			return fmt.Errorf("file '%s' already exists", s)
		}

		dir := filepath.Dir(s)
		if !FileExist(dir) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	cert := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"bp"},
			OrganizationalUnit: []string{"mt", "pd"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(100, 0, 0),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, h)
		}
	}

	var (
		caCert    *x509.Certificate
		caPrivKey *rsa.PrivateKey
	)

	if caKeyPath == "" { // 自签名
		caCert = &cert
		caPrivKey = privKey
		cert.IsCA = true
		cert.KeyUsage |= x509.KeyUsageCertSign
	} else {
		caPrivKey, err = loadPrivKeyFromPEM(caKeyPath)
		if err != nil {
			return err
		}

		caCert, err = loadCertFromPEM(caCertPath)
		if err != nil {
			return err
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &cert, caCert, &privKey.PublicKey, caPrivKey)
	if err != nil {
		return fmt.Errorf("create certificate: %v", err)
	}

	closeFile := func(f *os.File) {
		if e := f.Close(); e != nil && err == nil {
			err = e
		}
	}

	{
		certFile, err := os.OpenFile(outCert, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer closeFile(certFile)

		block := &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}
		if err := pem.Encode(certFile, block); err != nil {
			return fmt.Errorf("pem encode: %v", err)
		}
	}

	{
		keyFile, err := os.OpenFile(outKey, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer closeFile(keyFile)

		block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)}
		if err := pem.Encode(keyFile, block); err != nil {
			return err
		}
	}

	return nil
}
