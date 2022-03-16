// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"cloudiac/portal/libs/db"
	"cloudiac/utils"
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
	NewId() Id
}

type ModelIdSetter interface {
	SetId(interface{})
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

type AbstractModel struct {
}

func (AbstractModel) Migrate(*db.Session) error {
	return nil
}

func (AbstractModel) Validate() error {
	return nil
}

func (AbstractModel) ValidateAttrs(attrs Attrs) error {
	return nil
}

func (AbstractModel) AddUniqueIndex(sess *db.Session, index string, columns ...string) error {
	return sess.AddUniqueIndex(index, columns...)
}

type BaseModel struct {
	AbstractModel
	Id Id `gorm:"size:32;primary_key" json:"id" example:"x-c3ek0co6n88ldvq1n6ag"` //ID
}

func (base *BaseModel) SetId(id interface{}) {
	switch v := id.(type) {
	case string:
		base.Id = Id(v)
	case Id:
		base.Id = v
	default:
		panic(fmt.Errorf("invalid id type '%T'", id))
	}
}

func (base *BaseModel) CustomBeforeCreate(*db.Session) error {
	// 未设置 Id 值的情况下默认生成一个无前缀的 id，如果对前缀有要求请主动为对象设置 Id 值,
	// 或者在 Model 层定义自己的 CustomBeforeCreate() 方法
	if base.Id == "" {
		base.Id = NewId("")
	}
	return nil
}

// AutoUintIdModel  使用自增 uint id 的 model
type AutoUintIdModel struct {
	AbstractModel
	Id uint `gorm:"primary_key" json:"id"`
}

func (b *AutoUintIdModel) SetId(id interface{}) {
	switch v := id.(type) {
	case int:
		b.Id = uint(v)
	case uint:
		b.Id = v
	default:
		panic(fmt.Errorf("invalid id type '%T'", id))
	}
}

type TimedModel struct {
	BaseModel

	CreatedAt Time `json:"createdAt" gorm:"type:datetime" example:"2006-01-02 15:04:05"` // 创建时间
	UpdatedAt Time `json:"updatedAt" gorm:"type:datetime" example:"2006-01-02 15:04:05"` // 更新时间
}

type SoftDeleteModel struct {
	TimedModel
	DeletedAtT db.SoftDeletedAt `json:"deletedAt,omitempty" gorm:"default:0;not null;index" swaggerignore:"true"`
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

type Time time.Time

func (t *Time) UnmarshalJSON(bs []byte) error {
	tt, err := time.Parse(time.RFC3339, string(bs))
	if err != nil {
		return err
	}
	*t = Time(tt)
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))), nil
}

func (Time) Parse(s string) (t Time, err error) {
	if err = t.UnmarshalJSON([]byte(s)); err != nil {
		return t, err
	}
	return t, nil
}

// Value 获取时间值
// mysql 插入数据库的时候使用该函数
func (t Time) Value() (driver.Value, error) {
	var zeroTime time.Time
	if time.Time(t).UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return time.Time(t), nil
}

// Scan 转换为 time.Time
func (t *Time) Scan(v interface{}) error {
	switch value := v.(type) {
	case []byte:
		tv, err := time.Parse("2006-01-02 15:04:05", string(value))
		if err != nil {
			return err
		}
		*t = Time(tv)
	case time.Time:
		*t = Time(value)
	default:
		return fmt.Errorf("can not convert %v to timestamp", v)
	}
	return nil
}

func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}
