// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package ctx

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models/forms"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type RequestContext interface {
	BindService(sc *ServiceContext)
	Service() *ServiceContext
	Logger() logs.Logger
}

type GinRequest struct {
	*gin.Context
	sc   *ServiceContext
	form forms.BaseFormer
}

var (
	transZh ut.Translator
	transEn ut.Translator
)

func init() {
	transZh, transEn = register()
}

func NewGinRequest(c *gin.Context) *GinRequest {
	if rc, exist := c.Get(consts.CtxKey); exist {
		return rc.(*GinRequest)
	}

	ctx := &GinRequest{
		Context: c,
	}
	ctx.sc = NewServiceContext(ctx)
	c.Set(consts.CtxKey, ctx)
	return ctx
}

func (c *GinRequest) BindService(sc *ServiceContext) {
	c.sc = sc
}

func (c *GinRequest) Service() *ServiceContext {
	return c.sc
}

func (c *GinRequest) Logger() logs.Logger {
	return c.sc.Logger()
}

type JSONResult struct {
	Code           int               `json:"code" example:"200"`
	Message        string            `json:"message" example:"ok"`
	MessageDetail  string            `json:"message_detail,omitempty" example:"ok"`
	Result         interface{}       `json:"result,omitempty" swaggertype:"object"`
	ValidateErrors map[string]string `json:"validate_errors,omitempty"`
}

func (c *GinRequest) JSON(status int, msg interface{}, result interface{}) {
	var (
		message         = ""
		code            = 0
		detail          string
		validateErrors  map[string]string
		validationError validator.ValidationErrors
	)
	if msg != nil {
		if er, ok := msg.(e.Error); ok {

			if er.Status() != 0 {
				status = er.Status()
			}
			message = e.ErrorMsg(er, c.GetHeader("accept-language"))
			code = er.Code()
			lang := e.GetAcceptLanguage(c.GetHeader("accept-language"))
			if errors.As(msg.(e.Error).Err(), &validationError) {
				if lang == "en-US" {
					detail, validateErrors = setDetailAndValidateErrors(validationError, transEn)
				} else {
					detail, validateErrors = setDetailAndValidateErrors(validationError, transZh)
				}
			} else {
				detail = er.Error()
			}
		} else {
			code = e.InternalError
			message = fmt.Sprintf("%v", msg)
		}
	}

	if code != 0 {
		code, _ = strconv.Atoi(fmt.Sprintf("%d%05d", status, code))
	} else {
		code = status
	}

	jsonResult := JSONResult{
		Code:           code,
		Message:        message,
		MessageDetail:  detail,
		Result:         result,
		ValidateErrors: validateErrors,
	}

	c.Context.JSON(status, jsonResult)
}

func (c *GinRequest) JSONError(err e.Error, statusOrResult ...interface{}) {
	var (
		status = http.StatusInternalServerError
		result interface{}
	)
	for _, v := range statusOrResult {
		switch vv := v.(type) {
		case int:
			status = vv
		default:
			result = v
		}
	}

	c.JSON(status, err, result)
	c.Abort()
}

func (c *GinRequest) JSONSuccess(res ...interface{}) {
	if len(res) == 0 {
		c.JSON(http.StatusOK, nil, nil)
	} else {
		c.JSON(http.StatusOK, nil, res[0])
	}
	c.Abort()
}

func (c *GinRequest) JSONResult(res interface{}, err e.Error) {
	if err != nil {
		c.JSONError(err, res)
	} else {
		c.JSONSuccess(res)
	}
}

func (c *GinRequest) FileDownloadResponse(data []byte, filename string, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if filename != "" {
		c.Writer.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}
	c.Context.Data(http.StatusOK, contentType, data)
}

//BindUriTagOnly 将 context.Params 绑定到标记了 uri 标签的 form 字段
func BindUriTagOnly(c *GinRequest, b interface{}) error {
	if len(c.Params) == 0 {
		return nil
	}
	typs := reflect.TypeOf(b).Elem()
	vals := reflect.ValueOf(b).Elem()
	for _, p := range c.Params {
		for i := 0; i < typs.NumField(); i++ {
			if key, ok := typs.Field(i).Tag.Lookup("uri"); ok && reflect.ValueOf(p.Key).String() == key {
				v := reflect.ValueOf(p.Value)
				vals.Field(i).Set(v.Convert(vals.Field(i).Type()))
			}
		}
	}
	return nil
}

func (c *GinRequest) Bind(form forms.BaseFormer) error {

	var (
		jsonForm      map[string]interface{}
		err           error
		validateError validator.ValidationErrors
	)

	// 将 Params 绑定到 form 里面标记了 uri 的字段
	if err = BindUriTagOnly(c, form); err != nil {
		// URI 参数不对，按路径不对处理
		c.JSONError(e.New(e.BadParam, err), http.StatusNotFound)
		c.Abort()
		return err
	}

	// ShouldBind() 不支持 HTTP GET 请求通过 body json 传参，
	// 所以我们针对 json 类型的 content-type 做特殊处理
	if c.ContentType() == binding.MIMEJSON && c.Request.Method != "GET" {
		var body []byte
		body, err = ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSONError(e.New(e.IOError, err), http.StatusInternalServerError)
			c.Abort()
			return err
		}

		if len(body) > 0 { // body 有内容才不进行 json bind
			err = binding.JSON.BindBody(body, form)
			_ = json.Unmarshal(body, &jsonForm)
		} else { // 否则使用 Form binding
			err = c.Context.ShouldBindWith(form, binding.Form)
		}
	} else {
		err = c.Context.ShouldBind(form)
	}

	if err != nil {
		if errors.As(err, &validateError) {
			c.JSONError(e.New(e.BadParam, validateError), http.StatusBadRequest)

		} else {
			c.JSONError(e.New(e.BadParam, err), http.StatusBadRequest)
		}
		c.Abort()
		return err
	}

	if err := c.Request.ParseForm(); err != nil {
		c.JSONError(e.New(e.BadParam, err), http.StatusBadRequest)
		c.Abort()
		return err
	}

	values := url.Values{}
	// path 参数可以被 post 参数覆盖，需要先添加
	for _, p := range c.Params {
		values.Set(p.Key, p.Value)
	}
	// url 参数也可以被 post 参数覆盖
	for k, v := range c.Request.Form {
		values[k] = v
	}
	for k, v := range c.Request.PostForm {
		values[k] = v
	}
	for k, v := range jsonForm {
		values.Set(k, fmt.Sprintf("%v", v))
	}

	form.Bind(values)
	c.form = form

	return nil
}

func setDetailAndValidateErrors(validationError validator.ValidationErrors, trans ut.Translator) (string, map[string]string) {
	var data string
	validationErrors := make(map[string]string, len(validationError.Translate(trans)))
	valueArrays := make([]string, 0)
	for key, value := range validationError.Translate(trans) {
		newKey := key[strings.Index(key, ".")+1:]
		validationErrors[newKey] = value
		valueArrays = append(valueArrays, value)
	}
	data = strings.Join(valueArrays, "\n")
	return data, validationErrors
}

func register() (ut.Translator, ut.Translator) {
	f := func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())
		return t
	}
	validate := binding.Validator.Engine().(*validator.Validate)
	return zhRegisterTranslation(validate, f), enRegisterTranslation(validate, f)
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
