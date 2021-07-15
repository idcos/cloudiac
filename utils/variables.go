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

func DecryptSecretVar(value string) string {
	val, isSecret := DecodeSecretVar(value)
	if isSecret {
		return AesDecrypt(val)
	}
	return val
}
