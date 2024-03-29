// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

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

// 强制解密 value，不管是否带有加密前缀标识
func DecryptSecretVarForce(value string) (string, error) {
	// 先移除可能存在的加密前缀
	val, _ := DecodeSecretVar(value)
	// aes 解密
	return AesDecrypt(val)
}

// 加密字符串，并添加前缀标识
func EncryptSecretVar(value string) (string, error) {
	var err error
	if value, err = AesEncrypt(value); err != nil {
		return "", err
	}
	return EncodeSecretVar(value, true), nil
}
