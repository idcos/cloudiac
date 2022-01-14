// Copyright 2021 CloudJ Company Limited. All rights reserved.

package e

import (
	"cloudiac/utils/logs"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Error interface {
	Error() string
	Status() int
	Code() int
	Err() error
}

type MyError struct {
	httpStatus int
	code       int
	err        error
}

var logger = logs.Get()

func (e *MyError) Status() int {
	return e.httpStatus
}

func (e *MyError) Code() int {
	return e.code
}

func (e *MyError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return fmt.Sprintf("Error%d", e.code)
}

func (e *MyError) Err() error {
	return e.err
}

func newError(code int, err error, status int) Error {
	return &MyError{
		httpStatus: status,
		code:       code,
		err:        err,
	}
}

// New 生成一个 Error 对象，code 为错误码，errOrStatus 为错误消息或者 http status
// err 和 http status 可以同时传，函数自动根据数据类型来判断是哪种值。
// !!建议使用 AutoNew，可以自动判断 err 类型，如果 err 己是一个 Error 对象则不创建新 error!!
func New(code int, errOrStatus ...interface{}) Error {
	var (
		// 默认设置 http 状态码为 0，
		// GinRequestCtx.JSON() 函数只在 err.Status() 为非 0 时才会使用这里的状态码，否则使用 status 参数的值
		status       = 0
		err    error = nil
	)
	for _, es := range errOrStatus {
		switch v := es.(type) {
		case int:
			status = v
		case error:
			err = v
		default:
			logger.Warnf("invalid errOrStatus value type %T(%v)", es, es)
		}
	}

	coverCode := converVcsError(code, err)
	return convertError(coverCode, err, status)
}

func converVcsError(code int, err error) int {
	if code == VcsError {
		info := err.Error()
		switch {
		// 前面的是否包含后面的
		case strings.Contains(info, "unsupported protocol scheme"):
			// vcs地址错误
			return VcsAddressError
		case strings.Contains(info, "Unauthorized"):
			// vcs权限不足
			return VcsInvalidToken
		case strings.Contains(info, "connection refused"):
			// vcs连接失败
			return VcsConnectError
		case strings.Contains(info, "handshake failure"):
			// vcs 连接失败
			return VcsConnectError
		case strings.Contains(info, "timeout"):
			// vcs 连接超时
			return VcsConnectTimeOut
		}
	}
	return code

}

func convertError(code int, err error, status int) Error {
	switch code {
	case DBError:
		if e, ok := err.(*mysql.MySQLError); ok {
			switch e.Number {
			case MysqlDuplicate:
				return newError(ObjectAlreadyExists, err, status)
			case MysqlUnknownColumn:
				return newError(InvalidColumn, err, status)
			case MysqlDropColOrKeyNotExists:
			case MysqlTableNotExist:
				return newError(DBError, err, status)
			case MysqlDataTooLong:
				return newError(DataTooLong, err, status)
			}
		}
	}

	return newError(code, err, status)
}

func Is(err error, code int) bool {
	if er, ok := err.(Error); ok {
		return er.Code() == code
	}
	return false
}

func IsMysqlErr(err error, num int) bool {
	if e, ok := err.(*mysql.MySQLError); ok {
		if num == 0 {
			return true
		} else if e.Number == uint16(num) {
			return true
		}
		return false
	} else {
		return false
	}
}

func IsDuplicate(err error) bool {
	if er, ok := err.(*MyError); ok {
		err = er.Err()
	}
	return IsMysqlErr(err, MysqlDuplicate)
}

func IgnoreDuplicate(err error) error {
	if IsDuplicate(err) {
		return nil
	}
	return err
}

func IsRecordNotFound(err error) bool {
	if er, ok := err.(*MyError); ok {
		err = er.Err()
	}
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IgnoreNotFound(err error) error {
	if IsRecordNotFound(err) {
		return nil
	}
	return err
}

func GetErr(err error) (*MyError, bool) {
	er, ok := err.(*MyError)
	// logs.Get().Warnf("GetErr: %T: %v, %v", err, er, ok)
	return er, ok
}

func AutoNew(err error, code int, status ...int) Error {
	// 如果 err 是 Error 类型则直接返回
	if er, ok := GetErr(err); ok {
		return er
	}

	// 否则生成一个 code 对应的 Error
	if len(status) > 0 {
		return New(code, err, status[0])
	} else {
		return New(code, err)
	}
}

const defaultLang = "zh-cn"

func ErrorMsg(err Error, lang string) string {
	if lang == "" {
		lang = defaultLang
	}

	if m, ok := errorMsgs[err.Code()]; ok {
		if msg, ok := m[lang]; ok {
			return msg
		} else if msg, ok := m[defaultLang]; ok {
			return msg
		}
	}
	return err.Error()
}
