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

	txdb "github.com/DATA-DOG/go-txdb"
	"github.com/gin-gonic/gin"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/assert"
)

var (
	fixtures *testfixtures.Loader
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

//prepareMySQLDB 为测试用例 T 准备一个新的数据库连接
func prepareMySQLDB(t *testing.T) (sess *db.Session, cleanup func() error) {
	dsn := fmt.Sprintf("root:%s@tcp(localhost:%d)/iac_test", "Yunjikeji", 3307)
	// cName := fmt.Sprintf("tx_10") //, t.Name(), time.Now().UnixNano())
	err := db.InitWithTxdb(dsn, "mysql")
	if err != nil {
		t.Fatalf("open mysqltx connection: %s, err: %s", dsn, err)
	}

	close := func() error {
		sqlDb, err := db.Get().GormDB().DB()
		if err != nil {
			t.Fatalf("close db: %s", err)
		}
		return sqlDb.Close()
	}

	// 初始化数据库表
	// models.Init(true)

	sess = db.Get()
	sqlDb, _ := sess.GormDB().DB()

	// 初始化数据，这里会导入默认的管理员用户/组织/项目
	fixtures, err = testfixtures.New(
		testfixtures.Database(sqlDb),
		testfixtures.Dialect("mysql"),
		// 加载测试数据，可以使用 Paths, Directory, Files 这几种方式进行加载
		testfixtures.Paths(
			"../../unittest_init",
			"testdata/fixtures",
		),
	)
	if err != nil {
		t.Fatalf("load fixtures: %s", err)
	}

	return db.Get(), close
}

//prepareTestDatabase 加载测试数据，每次测试前都需要执行
func prepareTestDatabase(t *testing.T) (sess *db.Session, cleanup func() error) {
	fmt.Println("reload database =============")
	sess, cleanup = prepareMySQLDB(t)

	// 每次 load 都会清理旧数据并重新加载
	if err := fixtures.Load(); err != nil {
		t.Fatalf("load fixtures: %s", err)
	}
	return sess, cleanup
}

// TestMain 该函数在所有测试用例执行之前会被调用
func TestMain(m *testing.M) {
	fmt.Println("test main=============")
	txdb.Register("mysqltx", "mysql", fmt.Sprintf("root:%s@tcp(localhost:%d)/iac_test?charset=utf8mb4&parseTime=True&loc=Local", "Yunjikeji", 3307))

	// 初始化 config 的 jwt key，login 接口需要使用
	configs.Set(configs.Config{
		JwtSecretKey: "6xGzLKiX4dl0UE6aVuBGCmRWL7cQ+90W",
	})

	// 调用 os.Exit 开始测试，测试完成退出测试
	os.Exit(m.Run())
}

// func TestOperationLogMiddleware(t *testing.T) {
// 	t.Parallel()
// 	TestLogGet(t)
// 	TestLogPost(t)
// }

func TestLogPost(t *testing.T) {
	_, cleanup := prepareTestDatabase(t)
	// 如果测试过程中有创建临时文件，需要在测试结束的时候删除
	defer cleanup() // 测试完成关闭数据库连接，数据会被清理掉

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
	_, cleanup := prepareTestDatabase(t)
	// 如果测试过程中有创建临时文件，需要在测试结束的时候删除
	defer cleanup() // 测试完成关闭数据库连接，数据会被清理掉

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
