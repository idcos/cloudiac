package web

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/web/api/v1/handlers"
	"cloudiac/portal/web/middleware"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func performRequest(r http.Handler, method, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	bodyReader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, bodyReader)
	if method == "POST" {
		req.Header.Add("Content-Type", "application/json")
	}
	for header := range headers {
		req.Header.Add(header, headers[header])
	}
	w := httptest.NewRecorder() // http.ResponseWriter
	r.ServeHTTP(w, req)
	return w
}

// TestMain 该函数在所有测试用例执行之前会被调用
func TestMain(m *testing.M) {
	// 初始化 config 的 jwt key，login 接口需要使用
	configs.Set(&configs.Config{
		JwtSecretKey: "6xGzLKiX4dl0UE6aVuBGCmRWL7cQ+90W",
	})

	// 如果所有测试用例都用同一个数据库，可以在这里调用 LoadTestDatabase
	// _ = db.LoadTestDatabase(nil, []string{
	//	"../../unittest_init", // 基础数据
	//	"testdata/fixtures",   // 测试用例自定义数据
	// })

	// 调用 os.Exit 开始测试，测试完成退出测试
	os.Exit(m.Run())
}

func TestLogPost(t *testing.T) {
	_ = db.LoadTestDatabase(t, []string{
		"../../unittest_init",
		"testdata/fixtures",
	})

	// 初始化 gin
	w := ctrl.WrapHandler
	e := gin.New()

	// 本次测试的中间件，其他单元测试可以不引入
	e.Use(w(middleware.Operation))

	// 测试的路由，这里忽略了 rbac 验证的 ac() 函数，因为这里只是测试中间件，不需要验证权限
	e.POST("/api/v1/auth/login", w(handlers.Auth{}.Login))

	// 执行测试动作
	r := performRequest(e, "POST", "/api/v1/auth/login", `
	{
    "email": "admin@example.com",
    "password": "Yunjikeji#123"
}
	`, nil)

	// 检查结果
	assert.Equal(t, http.StatusOK, r.Code)
}

func TestLogGet(t *testing.T) {
	_ = db.LoadTestDatabase(t, []string{
		"../../unittest_init",
		"testdata/fixtures",
	})

	// 初始化 gin
	w := ctrl.WrapHandler
	e := gin.New()

	// 本次测试的中间件，其他单元测试可以不引入
	e.Use(w(middleware.Operation))

	// 一般都需要引入，在检查token有效性的同时解析了用户/组织/项目等信息，在大部分API中都需要
	e.Use(w(middleware.Auth))

	// 测试的路由
	e.GET("/api/v1/auth/me", w(handlers.Auth{}.GetUserByToken))

	// 执行测试动作
	r := performRequest(e, "GET", "/api/v1/auth/me", "", map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", getNewLoginToken(t)),
	})

	// 检查结果
	assert.Equal(t, http.StatusOK, r.Code)
	// 检查是否有操作日志记录
}

func getNewLoginToken(t *testing.T) string {
	// 生成登录授权
	token, err := services.GenerateToken(models.Id("u-c9cgu32s1s4c17dvmjug"), "admin@example.com", true, 1*24*time.Hour)
	if err != nil {
		t.Fatalf("generate token: %s", err)
	}
	return token
}
