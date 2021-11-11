// Copyright 2021 CloudJ Company Limited. All rights reserved.

package utils

import (
	"fmt"
	"strings"
)

const (
	SecretValuePrefix = "secret:"
)

func EncodeSecretVar(value string, isSecret bool) string {
	if isSecret {
		return fmt.Sprintf("%s%s", SecretValuePrefix, value)
	}
	return value
}

func DecodeSecretVar(value string) (string, bool) {
	if strings.HasPrefix(value, SecretValuePrefix) {
		return value[len(SecretValuePrefix):], true
	}
	return value, false
}

func DecryptSecretVar(value string) (string, error) {
	val, isSecret := DecodeSecretVar(value)
	if isSecret {
		return AesDecrypt(val)
	}
	return val, nil
}

func EncryptSecretVar(value string, isSecret bool) (string, error) {
	var err error
	if isSecret {
		if value, err = AesEncrypt(value); err != nil {
			return "", err
		}
	}
	return EncodeSecretVar(value, isSecret), nil
}
