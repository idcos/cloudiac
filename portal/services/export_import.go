package services

import (
	"cloudiac/configs"
	"cloudiac/utils"
)

// 对传入的字符串使用 exportSecretKey 加密，并添加加密前缀标识
// maybeSecret 为 true 时会尝试先解密字符串，否则直接将字符中当明文处理
func ExportSecretStr(s string, maybeSecret bool) string {
	if maybeSecret {
		var err error
		s, err = utils.DecryptSecretVar(s)
		if err != nil {
			panic(err)
		}
	}

	sk := configs.Get().ExportSecretKey
	es, err := utils.AesEncryptWithKey(s, sk)
	if err != nil {
		panic(err)
	}
	return utils.EncodeSecretVar(es, true)
}

// 传入加密后的值，使用 exportSecretKey 解密
// 如果 encryptAgain 为 true，会重新使用系统密钥加密字符串，并添加加密前缀标识
func ImportSecretStr(secretStr string, encryptAgain bool) (string, error) {
	sk := configs.Get().ExportSecretKey
	s, _ := utils.DecodeSecretVar(secretStr)
	s, err := utils.AesDecryptWithKey(s, sk)
	if err != nil {
		return "", err
	}
	if encryptAgain {
		return utils.EncryptSecretVar(s)
	}
	return s, nil
}

func ExportVariableValue(val string, senstive bool) string {
	if !senstive {
		return val
	}
	val, _ = utils.DecryptSecretVar(val)
	return ExportSecretStr(val, false)
}

func ImportVariableValue(val string, senstive bool) (string, error) {
	if !senstive {
		return val, nil
	}

	sk := configs.Get().ExportSecretKey
	s, _ := utils.DecodeSecretVar(val)
	s, err := utils.AesDecryptWithKey(s, sk)
	if err != nil {
		return "", err
	}
	return utils.AesEncrypt(s)
}
