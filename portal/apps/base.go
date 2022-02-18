// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/libs/ctx"
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

	// 处理排序
	// 注意: 如果 query 使用 Raw 构建查询语句，则 Order 排序不生效，
	// 需要 Raw 语句生成的时候手动调用 PageForm.OrderBy 构建排序语句
	query = form.Order(query)

	p := page.New(currentPage, pageSize, query)

	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	slice := reflect.MakeSlice(reflect.SliceOf(typ), 0, 0)
	slicePtr := reflect.New(slice.Type())
	slicePtr.Elem().Set(slice)
	result, err := p.Result(slicePtr.Interface())
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

func BaseHandler(c *ctx.ServiceContext, form *forms.BaseForm) (interface{}, e.Error) {
	c.AddLogField("action", "base")
	return nil, nil
}
