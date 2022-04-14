// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package middleware

import (
	"bytes"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"fmt"
	"io/ioutil"
	"regexp"
	"time"
)

type HandleTableAndDesc map[string]string

var defaultLang = "zh-CN"

const (
	MsgOperationLogCreate     = "创建了%s中%s的数据"
	MsgOperationLogCreateName = "名称为%s"
	MsgOperationLogUpdate     = "修改了%s中id为%s的数据"
	MsgOperationLogDelete     = "删除了%s中id为%s的数据"
)

var message = map[string]map[string]string{
	MsgOperationLogCreate: {
		"en-US": "Create in %s %s",
		"zh-CN": MsgOperationLogCreate,
	},
	MsgOperationLogCreateName: {
		"en-US": "with name %s",
		"zh-CN": MsgOperationLogCreateName,
	},
	MsgOperationLogUpdate: {
		"en-US": "Modify data in %s with id %s",
		"zh-CN": MsgOperationLogUpdate,
	},
	MsgOperationLogDelete: {
		"en-US": "Delete data in %s with id %s",
		"zh-CN": MsgOperationLogDelete,
	},
}

func getMsg(msgKey string, lang ...string) string {
	if len(lang) == 0 {
		return message[msgKey][defaultLang]
	}
	return message[msgKey][lang[0]]
}

// parseId 通过 RequestURI 解析资源名称
func parseId(requestURI string) string {
	// 请求 /api/v1/users/:userId
	// 匹配第四段的         ^^^^^^ userId
	regex := regexp.MustCompile("^/[^/]+/[^/]+/[^/]+/([^/?#]+)")
	match := regex.FindStringSubmatch(requestURI)
	if len(match) == 2 {
		return match[1]
	}

	return ""
}

func Operation(c *ctx.GinRequest) {
	opMethod := &OperationMethod{C: c}
	var opLog *models.OperationLog

	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Logger().Warnf("operation log read body err %v", err)
		return
	}
	// 恢复原始的 body
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// 执行其他中间件以及api路由
	c.Next()

	// 根据method方法名确定操作行为
	switch c.Request.Method {
	case "PUT":
		opLog, err = opMethod.putOperation(bodyBytes)
	case "POST":
		opLog, err = opMethod.postOperation(bodyBytes)
	case "DELETE":
		opLog, err = opMethod.deleteOperation()
	default:
		return
	}

	if err == nil {
		opLog.OperationStatus = c.Writer.Status()
		err = opLog.InsertLog()
		if err != nil {
			c.Logger().Warnf("insert operation log err %v", err)
		}
	}
}

type OperationMethod struct {
	C *ctx.GinRequest
}

func (o *OperationMethod) putOperation(bodyBytes []byte) (opLog *models.OperationLog, err error) {
	// 修改操作 需要记录修改的数据
	cxt := o.C.Service()
	url := o.C.Request.URL.String()
	operationLog := models.OperationLog{
		UserID:        cxt.UserId,
		Username:      cxt.Username,
		UserAddr:      cxt.UserIpAddr,
		OperationUrl:  url,
		OperationAt:   models.Time(time.Now()),
		OperationType: "Put",
		OperationInfo: fmt.Sprintf(getMsg(MsgOperationLogUpdate), parseRes(url), parseId(url)),
		Desc:          models.JSON(bodyBytes),
	}

	return &operationLog, nil
}

func (o *OperationMethod) postOperation(bodyBytes []byte) (opLog *models.OperationLog, err error) {
	// 新建操作 需要记录新插入的数据
	name := o.C.Request.Form.Get("name")
	if name != "" {
		name = fmt.Sprintf(getMsg(MsgOperationLogCreateName), name)
	}

	cxt := o.C.Service()
	url := o.C.Request.URL.String()
	operationLog := models.OperationLog{
		UserID:        cxt.UserId,
		Username:      cxt.Username,
		UserAddr:      cxt.UserIpAddr,
		OperationUrl:  url,
		OperationAt:   models.Time(time.Now()),
		OperationType: "Create",
		OperationInfo: fmt.Sprintf(getMsg(MsgOperationLogCreate), parseRes(url), name),
		Desc:          models.JSON(bodyBytes),
	}

	return &operationLog, nil
}

func (o *OperationMethod) deleteOperation() (opLog *models.OperationLog, err error) {
	// 删除操作
	cxt := o.C.Service()
	url := o.C.Request.URL.String()
	operationLog := models.OperationLog{
		UserID:        cxt.UserId,
		Username:      cxt.Username,
		UserAddr:      cxt.UserIpAddr,
		OperationUrl:  url,
		OperationAt:   models.Time(time.Now()),
		OperationType: "Delete",
		OperationInfo: fmt.Sprintf(getMsg(MsgOperationLogDelete), parseRes(url), parseId(url)),
	}

	return &operationLog, nil
}
