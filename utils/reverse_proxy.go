package utils

import (
	"cloudiac/utils/logs"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func ReverseProxy(api string, c *gin.Context) {

	u, err := url.Parse(api)
	if err != nil {
		panic(err)
	}

	director := func(req *http.Request) {
		logs.Get().Debugf("before redirect: %s %s", req.Method, req.URL.String())

		// req.Header = http.Header{}
		req.Header.Del("Authorization")
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.URL.Path = u.Path

		logs.Get().Debugf("after redirect: %s %s", req.Method, req.URL.String())
		// req.Header.Set("Content-Type", "application/json; application/x-www-form-urlencoded; charset=utf-8")
		// req.Header.Set("Content-Type", "multipart/form-data")
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}
