package ctx

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models/forms"
	"cloudiac/utils/logs"
	//"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	//"github.com/gocarina/gocsv"

	"github.com/gin-gonic/gin"
)

type GinRequestCtx struct {
	*gin.Context
	sc   *ServiceCtx
	form forms.Former
}

type SysRequestCtx struct {
	sc *ServiceCtx
}

func (s SysRequestCtx) BindServiceCtx(sc *ServiceCtx) {
	s.sc = sc
}

func (s SysRequestCtx) ServiceCtx() *ServiceCtx {
	return s.sc
}

func (s SysRequestCtx) Logger() logs.Logger {
	return s.sc.Logger()
}

func (c *GinRequestCtx) BindServiceCtx(sc *ServiceCtx) {
	c.sc = sc
}

func (c *GinRequestCtx) ServiceCtx() *ServiceCtx {
	return c.sc
}

func (c *GinRequestCtx) Logger() logs.Logger {
	return c.sc.Logger()
}

func NewRequestCtx(c *gin.Context) *GinRequestCtx {
	if rc, exist := c.Get("_request_ctx"); !exist {
		ctx := &GinRequestCtx{
			Context: c,
		}
		ctx.sc = NewServiceCtx(ctx)
		c.Set("_request_ctx", ctx)
		return ctx
	} else {
		return rc.(*GinRequestCtx)
	}
}

func NewSysRequestCtx() *SysRequestCtx {
	ctx := &SysRequestCtx{}
	ctx.sc = NewServiceCtx(ctx)
	return ctx
}

func convertMessage(i interface{}) string {
	if er, ok := i.(e.Error); ok {
		return er.Error()
	} else if er, ok := i.(error); ok {
		return er.Error()
	} else {
		return fmt.Sprintf("%v", i)
	}
}

func (c *GinRequestCtx) JSON(status int, msg interface{}, result interface{}) {
	var (
		message = ""
		code    = 0
		detail  interface{}
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
		code, _ = strconv.Atoi(fmt.Sprintf("%d%04d", status, code))
	} else {
		code = status
	}
	c.Context.JSON(status, gin.H{
		"code":           code,
		"message":        message,
		"message_detail": detail,
		"result":         result,
	})
}

func (c *GinRequestCtx) JSONError(err e.Error, statusOrResult ...interface{}) {
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

func (c *GinRequestCtx) JSONSuccess(res ...interface{}) {
	if len(res) == 0 {
		c.JSON(http.StatusOK, nil, nil)
	} else {
		c.JSON(http.StatusOK, nil, res[0])
	}
	c.Abort()
}

func (c *GinRequestCtx) JSONResult(res interface{}, err e.Error) {
	if err != nil {
		c.JSONError(err, res)
	} else {
		c.JSONSuccess(res)
	}
}

func (c *GinRequestCtx) JSONOpenError(err e.Error, statusOrResult ...interface{}) {
	c.JSONOpen(err, nil, "result")
	c.Abort()
}

func (c *GinRequestCtx) JSONOpen(msg interface{}, result interface{}, resultType string) {
	var (
		status  = "success"
		message = "处理成功"
	)

	if msg != nil {
		if er, ok := msg.(e.Error); ok {
			message = er.Error()
			status = "fail"
		}
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"status":   status,
		"message":  message,
		resultType: result,
	})
}

func (c *GinRequestCtx) JSONOpenSuccessItem(res ...interface{}) {
	if len(res) == 0 {
		c.JSONOpen(nil, nil, "item")
	} else {
		c.JSONOpen(nil, res[0], "item")
	}
	c.Abort()
}

func (c *GinRequestCtx) JSONOpenSuccessList(res ...interface{}) {
	if len(res) == 0 {
		c.JSONOpen(nil, nil, "list")
	} else {
		c.JSONOpen(nil, res[0], "list")
	}
	c.Abort()
}

func (c *GinRequestCtx) JSONOpenResultItem(res interface{}, err e.Error) {
	if err != nil {
		c.JSONOpenError(err, res)
	} else {
		c.JSONOpenSuccessItem(res)
	}
}

func (c *GinRequestCtx) JSONOpenResultList(res interface{}, err e.Error) {
	if err != nil {
		c.JSONOpenError(err, res)
	} else {
		c.JSONOpenSuccessList(res)
	}
}

func (c *GinRequestCtx) AbortIfError(err e.Error) bool {
	if err != nil {
		c.JSONError(err)
		return true
	}
	return false
}

func (c *GinRequestCtx) Bind(form forms.Former) error {
	var body []byte
	if c.ContentType() == "application/json" {
		body, _ = ioutil.ReadAll(c.Request.Body)
		// Write body back for ShouldBind() call
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(body)))
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

func (c *GinRequestCtx) AutoResult(res interface{}, err e.Error) {
	if err != nil {
		c.JSONError(err)
	} else {
		c.JSONSuccess(res)
	}
	return
}

func (c *GinRequestCtx) convertInt(s string) (int, bool) {
	val, err := strconv.Atoi(s)
	if err != nil {
		c.JSONError(e.New(e.BadRequest, err), http.StatusBadRequest)
		return 0, false
	}
	return val, true
}

func (c *GinRequestCtx) convertUint(s string) (uint, bool) {
	val, ok := c.convertInt(s)
	if !ok {
		return 0, ok
	}
	if val < 0 {
		c.JSONError(e.New(e.BadRequest), http.StatusBadRequest)
		return 0, false
	}
	return uint(val), ok
}

func (c *GinRequestCtx) QueryInt(key string) (int, bool) {
	return c.convertInt(c.Query(key))
}

func (c *GinRequestCtx) QueryUint(key string) (uint, bool) {
	return c.convertUint(c.Query(key))
}

func (c *GinRequestCtx) PostFormInt(key string) (int, bool) {
	return c.convertInt(c.PostForm(key))
}

func (c *GinRequestCtx) PostFormUint(key string) (uint, bool) {
	return c.convertUint(c.PostForm(key))
}
