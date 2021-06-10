package v1_test

import (
	"bytes"
	"cloudiac/configs"
	"cloudiac/libs/db"
	"cloudiac/services"
	"cloudiac/utils/logs"
	"cloudiac/web"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)
var (
	router = &gin.Engine{}
	token = ""
)


func init() {
	router = web.GetRouter()
	base,_ := os.Getwd()
	path := filepath.Dir(filepath.Dir(filepath.Dir(base)))
	configPath := path + "/config.yml"
	token, _ = services.GenerateToken(1, "yunji", true, 1*24*time.Hour)
	configs.Init(configPath)
	conf := configs.Get().Log
	logs.Init(conf.LogLevel, "", 0)
}

func addHeader(r *http.Request, token string) {
	r.Header.Add("Authorization", token)
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("IaC-Org-Id", "1")
}


func TestHelloWorld(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hello", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"code":2001000,"message":"","message_detail":null,"result":{"goos":"darwin","hello":"world"}}`, w.Body.String())

}


func GetIacResultMap(content []byte) map[string]interface{}{
	result := map[string]interface{}{}
	err := json.Unmarshal(content, &result)
	if err != nil {
		fmt.Println(err)
		return result
	}
	if result["result"] != nil {
		return result["result"].(map[string]interface{})
	}
	return result

}

//  闭环测试CRUD，测试完成清除创建的资源
func vcsGet()(int, map[string]interface{}) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/vcs/search?pageSize=10&currentPage=1", nil)
	req.Header.Add("Authorization",token)
	req.Header.Add("IaC-Org-Id", "1")
	router.ServeHTTP(w, req)
	resultMap := GetIacResultMap(w.Body.Bytes())

	return w.Code, resultMap
}


func vcsCreate() (result map[string]interface{},repCode int){
	vesBody := map[string]string{
		"name":"yunji",
		"vcsType": "gitlab",
		"address":"git@gitlab.idcos.com:cloudiac/cloudiac.git",
		"vcsToken": "c2b3cvbn8qhv2e96haq0",
		"status":"enable",
	}
	ves,_ := json.Marshal(&vesBody)
	reader := bytes.NewReader(ves)

	req, err := http.NewRequest("POST", "/api/v1/vcs/create", reader)
	if err != nil {
		fmt.Println(err)
		return
	}
	addHeader(req, token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resultMap := GetIacResultMap(w.Body.Bytes())
	return resultMap, w.Code

}




func vcsUpdate(id int) (int,map[string]interface{}) {
	vesBody := map[string]interface{} {
		"id":id,
		"status":"disable",
	}
	ves,_ := json.Marshal(&vesBody)
	reader := bytes.NewReader(ves)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/vcs/update", reader)
	addHeader(req, token)
	router.ServeHTTP(w, req)
	return w.Code, GetIacResultMap(w.Body.Bytes())

}

func vcsDelete(id int) int{
	vesBody := map[string]int{
		"id":id,
	}
	ves,_ := json.Marshal(&vesBody)
	reader := bytes.NewReader(ves)
	req, _ := http.NewRequest("DELETE", "/api/v1/vcs/delete", reader)
	addHeader(req, token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Code

}

func TestVcs(t *testing.T) {
	defer func() {
		db.Init()
		rows, _ := db.Get().Raw("truncate iac_vcs;").Rows()
		defer rows.Close()
	}()
	resultMap, repCode := vcsCreate()
	id := int(resultMap["id"].(float64))

	assert.Equal(t, 200, repCode)
	assert.Equal(t, "yunji", resultMap["name"])
	assert.Equal(t, "gitlab", resultMap["vcsType"])
	assert.Equal(t, "enable", resultMap["status"])
	assert.Equal(t, "git@gitlab.idcos.com:cloudiac/cloudiac.git", resultMap["address"])
	assert.Equal(t, "c2b3cvbn8qhv2e96haq0", resultMap["vcsToken"])
	code, result := vcsGet()
	assert.Equal(t, 200, code)
	assert.Equal(t, 1, int(result["total"].(float64)))
	assert.Equal(t, 10, int(result["pageSize"].(float64)))
	code, resultMap = vcsUpdate(id)
	assert.Equal(t, 200, code)
	assert.Equal(t, "yunji", resultMap["name"])
	assert.Equal(t, "gitlab", resultMap["vcsType"])
	assert.Equal(t, "disable", resultMap["status"])
	// TODO 测试删除不存在ID, 正常逻辑应当返回403
	//code = vcsDelete(9911)
	//fmt.Println(code)

	code = vcsDelete(id)
	assert.Equal(t, 200, code)


}
