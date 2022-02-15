// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package middleware

import (
	"cloudiac/portal/libs/ctx"
	"net/http"
)

var (
	allowHeaders  = "Content-Type,AccessToken,X-CSRF-Token,Authorization,Token"
	exposeHeaders = "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type"
)

func Cors(c *ctx.GinRequest) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", allowHeaders)
	c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS")
	c.Header("Access-Control-Expose-Headers", exposeHeaders)
	c.Header("Access-Control-Allow-Credentials", "true")

	//放行所有OPTIONS方法
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
	}

	c.Next()
}
