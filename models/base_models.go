package models

import (
	"time"

	"github.com/jinzhu/gorm"

	"cloudiac/libs/db"
)

type Attrs map[string]interface{}

type Modeler interface {
	TableName() string
	Validate() error
	ValidateAttrs(attrs Attrs) error
	Migrate(*db.Session) error
	//AddUniqueIndex(*db.Session) error
}

type BaseModel struct {
	Id uint `gorm:"primary_key" json:"id" csv:"id" tsdb:"-"`
}

func (BaseModel) Migrate(*db.Session) error {
	return nil
}

func (BaseModel) Validate() error {
	return nil
}

func (BaseModel) ValidateAttrs(attrs Attrs) error {
	return nil
}

func (BaseModel) AddUniqueIndex(sess *db.Session, index string, cols ...string) error {
	return sess.AddUniqueIndex(index, cols...)
}

type TimedModel struct {
	BaseModel

	CreatedAt time.Time `json:"createdAt" csv:"-" tsdb:"-"`
	UpdatedAt time.Time `json:"updatedAt" csv:"-" tsdb:"-"`
}

type SoftDeleteModel struct {
	TimedModel
	DeletedAt *time.Time `json:"-" csv:"-" sql:"index"`
	// 因为 deleted_at 字段的默认值为 NULL(gorm 也会依赖这个值做软删除)，会导致唯一约束与软删除冲突,
	// 所以我们增加 deleted_at_t 字段来避免这个情况。
	// 如果 model 需要同时支持软删除和唯一约束就需要在唯一约束索引中增加该字段
	// (使用 SoftDeleteModel.AddUniqueIndex() 方法添加索引时会自动加上该字段)。
	DeletedAtT int64 `json:"-" csv:"-" gorm:"default:0"`
}

func (SoftDeleteModel) AfterDelete(scope *gorm.Scope) error {
	if scope.Search.Unscoped {
		return nil
	}
	return scope.DB().Unscoped().UpdateColumn("deleted_at_t", time.Now().Unix()).Error
}

func (m SoftDeleteModel) AddUniqueIndex(sess *db.Session, index string, cols ...string) error {
	cols = append(cols, "deleted_at_t")
	return m.TimedModel.AddUniqueIndex(sess, index, cols...)
}

type IsolatedBaseModel struct {
	BaseModel
	OrgId uint `json:"orgId" gorm:"default:'0'"`
}

func (m IsolatedBaseModel) AddUniqueIndex(sess *db.Session, index string, cols ...string) error {
	cols = append(cols, "org_id")
	return m.BaseModel.AddUniqueIndex(sess, index, cols...)
}

type IsolatedTimedModel struct {
	TimedModel
	OrgId uint `json:"orgId" gorm:"default:'0'"`
}

func (m IsolatedTimedModel) AddUniqueIndex(sess *db.Session, index string, cols ...string) error {
	cols = append(cols, "org_id")
	return m.TimedModel.AddUniqueIndex(sess, index, cols...)
}

type IsolatedSoftDeleteModel struct {
	SoftDeleteModel

	OrgId uint
}

func (m IsolatedSoftDeleteModel) AddUniqueIndex(sess *db.Session, index string, cols ...string) error {
	cols = append(cols, "org_id")
	return m.SoftDeleteModel.AddUniqueIndex(sess, index, cols...)
}
