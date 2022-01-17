// Copyright 2021 CloudJ Company Limited. All rights reserved.

package page

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
)

type Paginator struct {
	Page   int // 当前页码(1 base)
	Size   int // 每页大小，为 0 表示不分页
	dbSess *db.Session
}

type PageResp struct {
	Total    int64       `json:"total" example:"1"`
	PageSize int         `json:"pageSize" example:"15"`
	List     interface{} `json:"list" swaggertype:"object"`
}

func New(page int, size int, q *db.Session) *Paginator {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = consts.DefaultPageSize
	} else if size > consts.MaxPageSize {
		size = consts.MaxPageSize
	}

	return &Paginator{
		Page:   page,
		Size:   size,
		dbSess: q,
	}
}

func (p *Paginator) MustTotal(outs ...interface{}) int64 {
	total, err := p.Total(outs...)
	if err != nil {
		panic(err)
	}
	return total
}

func (p *Paginator) Total(outs ...interface{}) (int64, error) {
	var (
		err   error
		count int64
	)

	if len(outs) > 0 {
		out := outs[0]
		count, err = p.dbSess.Model(out).Count()
	} else {
		count, err = p.dbSess.Count()
	}
	return count, err
}

func (p *Paginator) TotalBySubQuery() (int64, error) {
	var (
		err   error
		count int64
	)

	err = p.dbSess.Raw("SELECT COUNT(*) as count FROM (?) AS t", p.dbSess.Expr()).Row().Scan(&count)
	return count, err
}

func (p *Paginator) getPage() *db.Session {
	sess := p.dbSess.Limit(p.Size).Offset((p.Page - 1) * p.Size)
	// 数据分页时必须进行排序，如果查询未排序则默认使用 id 排序
	if !sess.IsOrdered() {
		sess = sess.Order("`id`")
	}
	return sess
}

func (p *Paginator) GetPage() *db.Session {
	return p.getPage()
}

func (p *Paginator) Scan(dest interface{}) error {
	return p.getPage().Scan(dest)
}

func (p *Paginator) Result(dest interface{}) (resp *PageResp, err error) {
	var (
		total int64
	)

	if total, err = p.TotalBySubQuery(); err != nil {
		return
	}

	if err = p.Scan(dest); err != nil {
		return nil, err
	}

	return &PageResp{
		Total:    total,
		PageSize: p.Size,
		List:     dest,
	}, nil
}

func (p *Paginator) Next() *Paginator {
	return &Paginator{
		Page:   p.Page + 1,
		Size:   p.Size,
		dbSess: p.dbSess,
	}
}
