package web

import (
	"cloudiac/configs"
	_ "cloudiac/docs"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/web/api/v1/handlers"
	"cloudiac/portal/web/middleware"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
)

func performRequest(r http.Handler, method, path string, body string) *httptest.ResponseRecorder {
	bodyReader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, bodyReader)
	if method == "POST" {
		req.Header.Add("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

const (
	userJson = `{
    "id": "u-c8q36oqs1s472lf4rdbg",
    "created_at": "2022-03-18 15:23:15",
    "updated_at": "2022-03-18 15:25:02",
    "deleted_at": null,
    "deleted_at_t": 0,
    "name": "admin",
    "email": "admin@example.com",
    "password": "$2a$10$3JhWsgA8OVIXTOydpq.SNutsVFu/qUlm69tK5V6ENHrQE8etMyu.a",
    "phone": "",
    "is_admin": 0,
    "status": "enable",
    "newbie_guide": null
  }`
)

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type Any struct{}

func (a Any) Match(v driver.Value) bool {
	return true
}

func userKeys() []string {
	m := make(map[string]interface{})
	json.Unmarshal([]byte(userJson), &m)
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func userVals() ([]string, []driver.Value) {
	u := models.User{}
	json.Unmarshal([]byte(userJson), &u)
	ps := reflect.ValueOf(u)
	fmt.Printf("ps: %v\n", ps)

	keys := userKeys()
	fmt.Printf("keys s: %v\n", keys)
	var vals []driver.Value
	for _, k := range keys {
		fmt.Printf("key %s\n", k)

		f := ps.FieldByName(strcase.ToCamel(k))

		if !f.IsValid() {
			vals = append(vals, nil)
		} else {
			vals = append(vals, f.Interface())
		}
		fmt.Printf("xxx")
		fmt.Printf("%s: %v\n", k, f)
		// vals = append(vals, f.Interface())
	}
	return keys, vals
}

func TestOperationLogMiddleware(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("init mockdb error %v", err)
	}
	defer mockDB.Close()

	db.InitMockDb(mockDB)

	// 初始化 config 的 jwt key，login 接口需要使用
	configs.Set(configs.Config{
		JwtSecretKey: "6xGzLKiX4dl0UE6aVuBGCmRWL7cQ+90W",
	})

	// 从 user json 转换为 key slice 和 value slice，作为后续查询记录返回使用
	keys, vals := userVals()
	// 单独处理 user.password,因为 user 的 password 字段被忽略
	for i, key := range keys {
		if key == "password" {
			vals[i] = "$2a$10$3JhWsgA8OVIXTOydpq.SNutsVFu/qUlm69tK5V6ENHrQE8etMyu.a"
		}
	}

	// mock 第一条数据库操作，返回 user 记录
	mock.ExpectQuery(`SELECT .*`). // 正则匹配查询语句 "SELECT * FROM `iac_user` WHERE email = ? AND `iac_user`.`deleted_at_t` = ? ORDER BY `iac_user`.`id` LIMIT 1"
		// 匹配 ? 查询参数
		WithArgs("admin@example.com", 0).
		WillReturnRows(
			// 返回结果 rows 的列名 []string
			sqlmock.NewRows(keys).
				// 返回查询记录 []driver.Value...
				AddRow(vals...))

	// mock 第二个数据库操作，插入一条操作记录
	mock.ExpectBegin()                                       // mock 开启事务
	mock.ExpectExec("INSERT INTO `iac_operation_log` (.+)"). // INSERT INTO `iac_operation_log` (`id`,`user_id`,`username`,`user_addr`,`operation_at`,`operation_url`,`operation_type`,`operation_info`,`operation_status`,`desc`) VALUES (?,?,?,?,?,?,?,?,?,?)
		// 按字段顺序填写 values 参数
		WithArgs(Any{}, Any{}, Any{}, Any{}, AnyTime{}, "/api/v1/auth/login", "Create", "创建了auth中的数据", 200, `
	{
    "email": "admin@example.com",
    "password": "Yunjikeji#123"
}
	`).
		// exec 操作返回参数 lastInsertId, affectedRows
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit() // mock 提交事务

	// 初始化 gin
	w := ctrl.WrapHandler
	e := gin.New()
	e.Use(w(middleware.Operation))
	e.POST("/api/v1/auth/login", w(handlers.Auth{}.Login))

	// 执行动作
	r := performRequest(e, "POST", "/api/v1/auth/login", `
	{
    "email": "admin@example.com",
    "password": "Yunjikeji#123"
}
	`)

	// 检查结果
	assert.Equal(t, http.StatusOK, r.Code)
}
