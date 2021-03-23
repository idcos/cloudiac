package services

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/patrickmn/go-cache"
	"net/http"
	"cloudiac/configs"
	"cloudiac/consts/e"
	"cloudiac/utils"
	"time"
)

var (
	c = cache.New(5*time.Minute, 10*time.Minute)
	SecretKey = "c1c3ik8rvdg331ivogcg"
)


type Data struct {
	Token string `json:"token"`
}

type Resp struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Result  UserInfo `json:"result"`
}

type UserInfo struct {
	Id                    uint    `json:"id"`
	AppId                 uint    `json:"appId"`
	OrgId                 uint    `json:"orgId"`
	Username              string `json:"username"`
	Enabled               bool   `json:"enabled"`
	AccountNonExpired     bool   `json:"accountNonExpired"`
	AccountNonLocked      bool   `json:"accountNonLocked"`
	CredentialsNonExpired bool   `json:"credentialsNonExpired"`
	Permissions     []Permission `json:"permissions"`
}

type Permission struct {
	Code    string   `json:"code"`
	Actions []string `json:"actions"`
}

func GetUserInfo(token string) (UserInfo, error) {
	v, found := c.Get(token)
	if !found {
		res, er := AuthTokenVerify(token)
		//fmt.Println(string(res))
		if er != nil {
			return UserInfo{}, er
		}
		var resp Resp
		_ = json.Unmarshal(res, &resp)
		if resp.Code != "200" {
			return UserInfo{}, e.New(e.ValidateError)
		}
		c.Set(token, resp, cache.DefaultExpiration)
		v = resp
	}
	resp, _ := (v).(Resp)
	return resp.Result, nil
}

func AuthTokenVerify(token string) ([]byte, error) {
	conf := configs.Get()
	data := Data{
		Token: token,
	}
	header := &http.Header{}
	header.Set("Content-Type", "application/json")

	userInfo, er := utils.HttpService(conf.Iam.Addr + conf.Iam.AuthApi,"POST",
		header, data, 5, 5)
	if er != nil {
		return nil, er
	}
	return userInfo, nil
}

type Claims struct {
	UserId   uint   `json:"userId"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
	jwt.StandardClaims
}

func GenerateToken(uid uint, name string, isAdmin bool, expireDuration time.Duration) (string, error) {
	expire := time.Now().Add(expireDuration)

	// 将 userId，姓名, 过期时间写入 token 中
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserId:  uid,
		Username: name,
		IsAdmin: isAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expire.Unix(),
		},
	})

	return token.SignedString([]byte(SecretKey))
}