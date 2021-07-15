package utils

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
)

// OpenSSHPublicKey 获取 private key 对应的 open ssh 格式的 public key
func OpenSSHPublicKey(privatePem []byte) ([]byte, error) {
	block, _ := pem.Decode(privatePem)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, err
	}

	pubKey := signer.PublicKey()
	b := &bytes.Buffer{}
	b.WriteString(pubKey.Type())
	b.WriteByte(' ')
	e := base64.NewEncoder(base64.StdEncoding, b)
	_, err = e.Write(pubKey.Marshal())
	if err != nil {
		return  nil, err
	}
	_ = e.Close()
	b.WriteString(" CloudIaC")
	return b.Bytes(), nil
}
