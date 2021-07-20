package middleware

import (
	"bytes"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"fmt"
	"io/ioutil"
	"time"

	//"github.com/gin-gonic/gin"
	"strings"
)

type HandleTableAndDesc map[string]string

func Operation(c *ctx.GinRequestCtx) {
	if c.Request.Method == "" {
		c.Next() // 注意 next()方法的作用是跳过该调用链去直接后面的中间件以及api路由
	}

	opMethod := &OperationMethod{C: c}
	// 根据method方法名确定操作行为
	switch c.Request.Method {
	case "PUT":
		err := opMethod.putOperation()
		if err != nil {
			c.Next()
		}
	case "POST":
		err := opMethod.postOperation()
		if err != nil {
			c.Next()
		}
	case "DELETE":
		err := opMethod.deleteOperation()
		if err != nil {
			c.Next()
		}
	default:
		return
	}

}

type OperationMethod struct {
	C *ctx.GinRequestCtx
}

func (o *OperationMethod) putOperation() (err error) {
	bodyBytes, err := ioutil.ReadAll(o.C.Request.Body)
	o.C.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	cxt := o.C.ServiceCtx()
	url := o.C.Request.URL.String()
	operationLog := models.OperationLog{
		UserID:        cxt.UserId,
		Username:      cxt.Username,
		UserAddr:      cxt.UserIpAddr,
		OperationAt:   models.Time(time.Now()),
		OperationType: "PUT",
		OperationInfo: fmt.Sprintf("修改了%s中的数据", strings.Split(url, "/")[len(strings.Split(url, "/"))-2]),
		Desc:          models.JSON(bodyBytes),
	}

	err = operationLog.InsertLog()
	if err != nil {
		return err
	}

	return
}

func (o *OperationMethod) postOperation() (err error) {
	// 新建操作 需要记录新插入的数据
	name := o.C.PostForm("name")
	bodyBytes, err := ioutil.ReadAll(o.C.Request.Body)
	o.C.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	cxt := o.C.ServiceCtx()
	url := o.C.Request.URL.String()
	operationLog := models.OperationLog{
		UserID:        cxt.UserId,
		Username:      cxt.Username,
		UserAddr:      cxt.UserIpAddr,
		OperationAt:   models.Time(time.Now()),
		OperationType: "Create",
		OperationInfo: fmt.Sprintf("创建了%s中的名称为%s数据", strings.Split(url, "/")[len(strings.Split(url, "/"))-2], name),
		Desc:          models.JSON(bodyBytes),
	}

	err = operationLog.InsertLog()
	if err != nil {
		return
	}

	return
}

func (o *OperationMethod) deleteOperation() (err error) {
	// 删除操作
	cxt := o.C.ServiceCtx()
	id := o.C.PostForm("id")
	bodyBytes, err := ioutil.ReadAll(o.C.Request.Body)
	o.C.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	url := o.C.Request.URL.String()
	operationLog := models.OperationLog{
		UserID:        cxt.UserId,
		Username:      cxt.Username,
		UserAddr:      cxt.UserIpAddr,
		OperationAt:   models.Time(time.Now()),
		OperationType: "delete",
		OperationInfo: fmt.Sprintf("删除%s中了id为%s的数据", strings.Split(url, "/")[len(strings.Split(url, "/"))-2], id),
		Desc:          models.JSON(bodyBytes),
	}

	err = operationLog.InsertLog()
	if err != nil {
		return
	}

	return nil
}
