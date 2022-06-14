// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package validate

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"strings"
)

var (
	TransZh, TransEn = initRegisterTranslation()
	validate         *validator.Validate
)

func SetDetailAndValidateErrors(validationError validator.ValidationErrors, trans ut.Translator) (string, map[string]string) {
	validationErrors := make(map[string]string, len(validationError.Translate(trans)))
	valueArrays := make([]string, 0)

	for key, value := range validationError.Translate(trans) {
		newKey := key[strings.Index(key, ".")+1:]
		validationErrors[newKey] = value
		valueArrays = append(valueArrays, value)
	}

	return strings.Join(valueArrays, "\n"), validationErrors
}

var registerTranslationsFunc = func(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())
	return t
}

func initRegisterTranslation() (ut.Translator, ut.Translator) {
	return zhRegisterTranslation(GetValidate(), registerTranslationsFunc), enRegisterTranslation(GetValidate(), registerTranslationsFunc)
}

func initValidate() *validator.Validate {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		panic("bind init error")
	}

	return v
}

func GetValidate() *validator.Validate {
	if validate == nil {
		return initValidate()
	}
	return validate
}

func RegisterValida() {
	v := GetValidate()
	// 自定义bind参数校验tag
	if err := v.RegisterValidation("keysstartswith", func(fl validator.FieldLevel) bool {
		if strings.HasPrefix(fl.Field().String(), "-----BEGIN OPENSSH PRIVATE KEY-----") ||
			strings.HasPrefix(fl.Field().String(), "-----BEGIN RSA PRIVATE KEY-----") {
			return true
		}
		return false
	}); err != nil {
		panic(err)
	}
}

func zhRegisterTranslation(validate *validator.Validate, f func(ut ut.Translator, fe validator.FieldError) string) ut.Translator {
	translatorZh := zh.New()
	uniZh := ut.New(translatorZh, translatorZh)
	tranZh, _ := uniZh.GetTranslator("zh")
	err := zh_translations.RegisterDefaultTranslations(validate, tranZh)
	if err != nil {
		panic(err)
	}
	re(validate, tranZh, "required_with", registrationFunc("required_with", "{0}和{1}必须同时存在", false), f)
	re(validate, tranZh, "required_without", registrationFunc("required_without", "{0}必须存在如果{1}不存在", false), f)
	re(validate, tranZh, "required_without_all", registrationFunc("required_without_all", "{0}和[{1}]必须存在其中一种", false), f)
	re(validate, tranZh, "file", registrationFunc("file", "{0}必须为文件路径", false), f)
	return tranZh
}

func enRegisterTranslation(validate *validator.Validate, f func(ut ut.Translator, fe validator.FieldError) string) ut.Translator {
	translatorEn := en.New()
	uniEn := ut.New(translatorEn, translatorEn)
	tranEn, _ := uniEn.GetTranslator("en")
	err := en_translations.RegisterDefaultTranslations(validate, tranEn)
	if err != nil {
		panic(err)
	}
	re(validate, tranEn, "startswith", registrationFunc("startswith", "{0} must startswith '{1}'", false), f)
	re(validate, tranEn, "required_with", registrationFunc("required_with", "{0} and {1} must be also provided", false), f)
	re(validate, tranEn, "required_without", registrationFunc("required_without", "{0} must be provided if {1} is empty", false), f)
	re(validate, tranEn, "required_without_all", registrationFunc("required_without_all", "either provide {0} or provide [{1}]", false), f)
	re(validate, tranEn, "file", registrationFunc("file", "{0} must be a valid file", false), f)
	return tranEn
}

func re(v *validator.Validate, trans ut.Translator, tag string, customRegisFunc validator.RegisterTranslationsFunc, customTransFunc validator.TranslationFunc) {
	if err := v.RegisterTranslation(tag, trans, customRegisFunc, customTransFunc); err != nil {
		panic(err)
	}
}

func registrationFunc(tag string, translation string, override bool) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) (err error) {
		if err = ut.Add(tag, translation, override); err != nil {
			return
		}
		return
	}
}
