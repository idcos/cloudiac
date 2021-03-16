package models

import (
	"fmt"
	"cloudiac/libs/db"
	"time"
)

const (
	UserStatusUnknown  = 0
	UserStatusInactive = 1
	UserStatusNormal   = 2
	UserStatusDisabled = 3
)

type User struct {
	SoftDeleteModel

	Username  string    `json:"username" gorm:"size:64;not null;comment:'姓名'"`
	Email     string    `json:"email" gorm:"comment:'邮箱'"`
	Password  string    `json:"-" gorm:"not null;comment:'密码'"`
	Phone     string    `json:"phone" gorm:"comment:'电话'"`
	LastLogin time.Time `json:"lastLogin" gorm:"default:NULL;comment:'最后登录时间'"`
	Ldap      int       `json:"-" gorm:"default:0;comment:'ldap用户'"` // 1.ldap用户

	Status         int `json:"status" gorm:"type:tinyint;default:1;comment:'状态'"`
	LoginFailedCnt int `json:"-" gorm:"default:0;comment:'登录失败次数'"`

	// 默认组织 id(每个用户都有一个默认组织，只有这个组织的管理员可以对其禁用和重置密码)
	DefaultTid uint `json:"defaultTid" gorm:"default:0;comment:'默认组织'"`
}

func (User) TableName() string {
	return "t_user"
}

func (u *User) Validate() error {
	attrs := Attrs{
		"username": u.Username,
		"phone":    u.Phone,
		"email":    u.Email,
	}
	return u.ValidateAttrs(attrs)
}

func (u User) ValidateAttrs(attrs Attrs) error {
	for k, v := range attrs {
		name := db.ToColName(k)
		switch name {
		case "username":
			if v == "" {
				return fmt.Errorf("blank user name")
			}
		}
	}

	return nil
}

func (u User) Migrate(sess *db.Session) (err error) {
	err = u.AddUniqueIndex(sess, "unique__email", "email")
	if err != nil {
		return err
	}

	return nil
}
