// Copyright 2021 CloudJ Company Limited. All rights reserved.

package utils

import (
	"fmt"
	"strings"
)

const (
	SecretValuePrefix = "secret:"
)

// 为加密字符串添加前缀标识
func EncodeSecretVar(value string, isSecret bool) string {
	if isSecret {
		fmt.Println(2222)
		return fmt.Sprintf("%s%s", SecretValuePrefix, value)
	}
	return value
}

// 移除加密字符串的前缀标识
func DecodeSecretVar(value string) (string, bool) {
	if strings.HasPrefix(value, SecretValuePrefix) {
		return value[len(SecretValuePrefix):], true
	}
	return value, false
}

// 如果字符串有加密标识则解密，否则直接返回字符串
func DecryptSecretVar(value string) (string, error) {
	val, isSecret := DecodeSecretVar(value)
	if isSecret {
		return AesDecrypt(val)
	}
	return val, nil
}

// 加密字符串，并添加前缀标识
func EncryptSecretVar(value string) (string, error) {
	var err error
	if value, err = AesEncrypt(value); err != nil {
		fmt.Println(err, "1111")
		return "", err
	}
	return EncodeSecretVar(value, true), nil
}
