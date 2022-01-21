package services

import (
	"cloudiac/configs"
	"cloudiac/utils"
	"testing"
)

func TestExportVariatbleValue(t *testing.T) {
	configs.Set(configs.Config{
		SecretKey:       utils.Md5String("secretKey"),
		ExportSecretKey: utils.Md5String("exportSecretKey"),
	})

	val := "value"

	s, err := utils.AesEncrypt(val)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("value encrypte as '%s'", s)

	isSenstive := true
	es := ExportVariableValue(s, isSenstive)
	t.Logf("export variable '%s' as '%s'", val, es)

	is, err := ImportVariableValue(es, isSenstive)
	if err != nil {
		t.Fatal(err)
	}

	iVal, err := utils.AesDecrypt(is)
	if err != nil {
		t.Fatal(err)
	}

	if iVal != val {
		t.Fatalf("import variable error, except '%s', got '%s'", val, iVal)
	}
}
