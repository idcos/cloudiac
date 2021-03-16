package apps

import (
	"reflect"

	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/libs/page"
	"cloudiac/models/forms"
)

type TableIface interface {
	TableName() string
}

// 传入的 model 必须保证有 TableName() 方法，以确保可以获得要查询的表名
func getPage(query *db.Session, form forms.Former, model TableIface) (interface{}, e.Error) {
	pageSize := form.PageSize()
	currentPage := form.CurrentPage()
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

func getEmptyListResult(form forms.Former) (interface{}, e.Error) {
	pageSize := form.PageSize()
	currentPage := form.CurrentPage()
	p := page.New(currentPage, pageSize, nil)

	return page.PageResp{
		Total:    0,
		PageSize: p.Size,
		List:     []string{},
	}, nil
}
