package e

import (
	"cloudiac/utils/logs"
	"fmt"
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

// 生成一个 Error 对象
// code 为错误码
// errOrStatus 为错误消息或者 http status，可以同时传两者，自动根据数据类型来判断是哪种值
func New(code int, errOrStatus ...interface{}) Error {
	var (
		// 默认设置 http 状态码为 0，
		// GinRequestCtx.JSON() 函数只在 err.Status() 为非 0 时才会使用这里的状态码，否则使用 status 参数的值
		status       = 0
		err    error = nil
	)
	for _, v := range errOrStatus {
		switch v.(type) {
		case int:
			status = v.(int)
		case error:
			err = v.(error)
		default:
			logger.Errorf("'msgOrStatus' only supports 'string' or 'error'")
		}
	}

	return convertError(code, err, status)
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
			}
		}
	}

	return newError(code, err, status)
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

func AutoNew(err error, code int) Error {
	// 如果 err 是 Error 类型则直接返回
	if er, ok := GetErr(err); ok {
		return er
	}
	// 否则生成一个 code 对应的 Error
	return New(code, err)
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
