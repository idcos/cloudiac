package apps

import (
	"cloudiac/portal/libs/ctx"
	"fmt"
	"reflect"

	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models/forms"
)

type TableIface interface {
	TableName() string
}

// 传入的 model 必须保证有 TableName() 方法，以确保可以获得要查询的表名
func getPage(query *db.Session, form forms.PageFormer, model TableIface) (interface{}, e.Error) {
	pageSize := form.PageSize()
	currentPage := form.CurrentPage()
	query = form.Order(query)
	p := page.New(currentPage, pageSize, query)

	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	models := reflect.New(reflect.SliceOf(typ)).Interface()
	result, err := p.Result(models)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return result, nil
}

func getEmptyListResult(form forms.PageFormer) (interface{}, e.Error) {
	pageSize := form.PageSize()
	currentPage := form.CurrentPage()
	p := page.New(currentPage, pageSize, nil)

	return page.PageResp{
		Total:    0,
		PageSize: p.Size,
		List:     []string{},
	}, nil
}

func BaseHandler(c *ctx.ServiceCtx, form *forms.BaseForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("base"))
	return nil, nil
}
