package ctx

import (
	"bytes"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models/forms"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
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
	Code          int         `json:"code" example:"200"`
	Message       string      `json:"message" example:"ok"`
	MessageDetail string      `json:"message_detail,omitempty" example:"ok"`
	Result        interface{} `json:"result,omitempty" swaggertype:"object"`
}

func (c *GinRequest) JSON(status int, msg interface{}, result interface{}) {
	var (
		message = ""
		code    = 0
		detail  string
	)

	if msg != nil {
		if er, ok := msg.(e.Error); ok {
			if er.Status() != 0 {
				status = er.Status()
			}
			message = e.ErrorMsg(er, c.GetHeader("accept-language"))
			code = er.Code()
			detail = er.Error()
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
		Code:          code,
		Message:       message,
		MessageDetail: detail,
		Result:        result,
	}

	c.Context.JSON(status, jsonResult)
}

func (c *GinRequest) JSONError(err e.Error, statusOrResult ...interface{}) {
	var (
		status = http.StatusInternalServerError
		result interface{}
	)
	for _, v := range statusOrResult {
		switch v.(type) {
		case int:
			status = v.(int)
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
	var body []byte
	if c.ContentType() == "application/json" {
		body, _ = ioutil.ReadAll(c.Request.Body)
		// Write body back for ShouldBind() call
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(body)))
	}

	// 将 Params 绑定到 form 里面标记了 uri 的字段
	if err := BindUriTagOnly(c, form); err != nil {
		// URI 参数不对，按路径不对处理
		c.Logger().Errorf("bind uri error %s", err)
		c.JSON(http.StatusNotFound, e.New(e.BadParam, err), nil)
		c.Abort()
		return err
	}

	if err := c.Context.ShouldBind(form); err != nil {
		c.JSON(http.StatusBadRequest, e.New(e.BadParam, err), nil)
		c.Abort()
		return err
	}

	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, e.New(e.BadParam, err), nil)
		c.Abort()
		return err
	}

	values := url.Values{}
	// path 参数可以被 post 参数覆盖
	for _, p := range c.Params {
		values[p.Key] = []string{fmt.Sprintf("%v", p.Value)}
	}
	for k, v := range c.Request.Form {
		values[k] = v
	}
	for k, v := range c.Request.PostForm {
		values[k] = v
	}
	if c.ContentType() == "application/json" {
		var jsObj map[string]interface{}
		_ = json.Unmarshal(body, &jsObj)
		for k := range jsObj {
			values[k] = []string{fmt.Sprintf("%v", jsObj[k])}
		}
	}

	form.Bind(values)
	c.form = form

	return nil
}
