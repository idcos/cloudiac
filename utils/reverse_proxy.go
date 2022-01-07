package utils

import (
	"cloudiac/utils/logs"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func ReverseProxy(api string) gin.HandlerFunc {

	u, err := url.Parse(api)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			req.Header = http.Header{}
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.URL.Path = u.Path

			logs.Get().Debugf("redirect to registry: %s %s", req.Method, req.URL.String())
			// req.Header.Set("Content-Type", "application/json; charset=utf-8")
		}

		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
