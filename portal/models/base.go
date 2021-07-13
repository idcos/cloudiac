package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"cloudiac/portal/libs/db"
	"cloudiac/utils"
	"github.com/jinzhu/gorm"
)

type Attrs map[string]interface{}

const (
	Enable  = "enable"
	Disable = "disable"
)

type Modeler interface {
	TableName() string
	Validate() error
	ValidateAttrs(attrs Attrs) error
	Migrate(*db.Session) error
	//AddUniqueIndex(*db.Session) error
}

type ModelIdGenerator interface {
	NewId() string
}

type Id string

func NewId(prefix string) Id {
	return Id(utils.GenGuid(prefix))
}

// InArray 检查 id 是否在数组中
func (i *Id) InArray(arr ...Id) bool {
	for idx := range arr {
		if arr[idx] == *i {
			return true
		}
	}
	return false
}

func (i Id) Value() (driver.Value, error) {
	return string(i), nil
}

func (i *Id) Scan(value interface{}) error {
	*i = ""
	switch v := value.(type) {
	case []byte:
		*i = Id(v)
	case string:
		*i = Id(v)
	default:
		return fmt.Errorf("invalid type %T, value: %T", value, value)
	}
	return nil
}

func (i Id) String() string {
	return string(i)
}

type BaseModel struct {
	Id Id `gorm:"size:32;primary_key" json:"id" example:"x-c3ek0co6n88ldvq1n6ag"` //ID
}

func (base *BaseModel) BeforeCreate(scope *gorm.Scope) error {
	// 未设置 Id 值的情况下默认生成一个无前缀的 id，如果对前缀有要求请主动为对象设置 Id 值,
	// 或者在 Model 层定义自己的 BeforeCreate() 方法
	if base.Id == "" {
		base.Id = NewId("")
	}
	return nil
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

// AutoUintIdModel  使用自增 uint id 的 model
type AutoUintIdModel struct {
	BaseModel
	//Id uint `gorm:"primary_key"`
}

type TimedModel struct {
	BaseModel

	CreatedAt utils.JSONTime `json:"createdAt" gorm:"type:datetime" example:"2006-01-02 15:04:05"` // 创建时间
	UpdatedAt utils.JSONTime `json:"updatedAt" gorm:"type:datetime" example:"2006-01-02 15:04:05"` // 更新时间
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

type StrSlice []string

func (v StrSlice) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *StrSlice) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}
