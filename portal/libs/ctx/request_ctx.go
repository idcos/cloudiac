package ctx

import (
	"cloudiac/utils/logs"
)

type RequestContextInter interface {
	BindServiceCtx(sc *ServiceCtx)
	ServiceCtx() *ServiceCtx
	Logger() logs.Logger
}
