package sshkey

import (
	"cloudiac/configs"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"os"
)

func generateSSHKeyPair(privatePath string, publicPath string) (err error) {
	openFile := func(p string) (*os.File, error) {
		if _, err := os.Stat(p); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			return nil, errors.Wrap(os.ErrExist, p)
		}
		return os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0600)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	privateFile, err := openFile(privatePath)
	if err != nil {
		return err
	}
	defer privateFile.Close()

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err = pem.Encode(privateFile, privateKeyPEM); err != nil {
		return err
	}

	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	publicFile, err := openFile(publicPath)
	if err != nil {
		return err
	}
	defer publicFile.Close()

	_, err = publicFile.Write(ssh.MarshalAuthorizedKey(sshPublicKey))
	return err
}

func LoadPrivateKeyPem() ([]byte, error) {
	c := configs.Get().Portal
	return os.ReadFile(c.SSHPrivateKey)
}

func InitSSHKeyPair() error {
	c := configs.Get().Portal
	if utils.FileExist(c.SSHPrivateKey) {
		return nil
	}

	logs.Get().Infof("generate ssh key pair: %s", c.SSHPrivateKey)
	return generateSSHKeyPair(c.SSHPrivateKey, c.SSHPublicKey)
}
