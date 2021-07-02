package forms

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/utils/logs"
	"fmt"
	"net/url"
	"strings"
)

type BaseFormer interface {
	Bind(url.Values)
	Get(string) (string, bool)
	HasKey(string) bool
}

type PageFormer interface {
	BaseFormer

	CurrentPage() int
	PageSize() int
	Export() bool
	Order(*db.Session) *db.Session
}

type BaseForm struct {
	formValues url.Values
}

type PageForm struct {
	BaseForm

	PageSize_    int  `form:"pageSize" json:"pageSize" binding:""`       // 分页大小，默认 15 条
	CurrentPage_ int  `form:"currentPage" json:"currentPage" binding:""` // 当前分页，从 1 开始
	Export_      bool `form:"export" json:"export" binding:""`           // 是否导出 csv 文件

	SortField_ string `form:"sortField" json:"sortField"`                  // 排序字段名称
	SortOrder_ string `form:"sortOrder" json:"sortOrder" enums:"asc,desc"` // 排序顺序
}

func (b *BaseForm) Bind(values url.Values) {
	b.formValues = values
}

func (b *BaseForm) Get(k string) (string, bool) {
	values, ok := b.formValues[k]
	if !ok {
		return "", false
	}
	if len(values) == 0 {
		return "", true
	}
	return values[0], true
}

func (b *BaseForm) HasKey(k string) bool {
	_, ok := b.Get(k)
	return ok
}

func (b *PageForm) CurrentPage() int {
	if b.Export_ {
		// 导出 csv 总是从第一页开始
		return 1
	}
	if b.CurrentPage_ <= 0 {
		return 1
	}
	return b.CurrentPage_
}

func (b *PageForm) PageSize() int {
	if b.Export_ {
		// 导出 csv 时设定为最大单页数量
		return consts.MaxPageSize
	}
	if b.PageSize_ > consts.MaxPageSize {
		return consts.MaxPageSize
	}
	if b.PageSize_ <= 0 {
		return consts.DefaultPageSize
	}
	return b.PageSize_
}

func (b *PageForm) Export() bool {
	return b.Export_
}

func (b *PageForm) SortField() string {
	return db.ToColName(b.SortField_)
}

func (b *PageForm) SortOrder() string {
	switch b.SortOrder_ {
	case "asc", "ascending":
		return "asc"
	case "desc", "descending":
		return "desc"
	default:
		return ""
	}
}

func (b *PageForm) Order(query *db.Session) *db.Session {
	if b.SortField() == "" {
		return query
	}

	if strings.Contains(b.SortField(), "`") {
		logs.Get().Warnf("invalid sortField: %s", b.SortField())
		return query
	}

	if b.SortOrder() == "desc" {
		return query.Order(fmt.Sprintf("`%s` desc", b.SortField()))
	} else {
		return query.Order(fmt.Sprintf("`%s`", b.SortField()))
	}
}
