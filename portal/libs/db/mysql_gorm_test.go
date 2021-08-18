package db

import (
	"fmt"
	"testing"
)

func TestInsert(t *testing.T) {
	mysqlConfig := "root:123456@tcp(127.0.0.1:3306)/iac?charset=utf8mb4&parseTime=True&loc=Local"
	Init(mysqlConfig)
	type Project struct {
		Id          string
		OrgId       string `json:"orgId" gorm:"size:32;not null"`     //组织ID
		Name        string `json:"name" form:"name" gorm:"not null;"` //组织名称
		Description string `json:"description" gorm:"type:text"`      //组织详情
		CreatorId   string `json:"creatorId" form:"creatorId" `       //用户id
		Status      string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:状态"`
	}

	m := []Project{
		Project{Id: "A", Name: "1"},
		Project{Id: "B", Name: "12"},
		Project{Id: "C", Name: "13"},
		Project{Id: "D", Name: "14"},
		Project{Id: "E", Name: "15"},
		Project{Id: "F", Name: "16"},
		Project{Id: "G", Name: "17"},
		Project{Id: "H", Name: "18"},
	}
	db := Get().Table("iac_project")
	err := db.Insert(m)
	fmt.Println(fmt.Sprintf("insert err: %v", err))
	//count, err := db.Delete(m)
	//fmt.Println(fmt.Sprintf("delete count %d , delete err: %v", count, err))
}
