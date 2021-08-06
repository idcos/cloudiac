// Copyright 2021 CloudJ Company Limited. All rights reserved.

package sshkey

/*
ssh 密钥生成和加载。
旧版本中是使用平台统一生成的 ssh key 并添加到部署的时候，新版本中改为由用户添加 ssh key
*/

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
	"path/filepath"
)

func generateSSHKeyPair(privatePath string, publicPath string) (err error) {
	openFile := func(p string) (*os.File, error) {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return nil, err
		}

		if _, err := os.Stat(p); err == nil {
			return nil, errors.Wrap(os.ErrExist, p)
		} else if os.IsNotExist(err) {
			return os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0600)
		} else {
			return nil, err
		}
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
