// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handler

import (
	"cloudiac/configs"
	"cloudiac/runner"
	"cloudiac/runner/api/ctx"
	"net/http"
	"path/filepath"
	"strings"
)

func RunClearProviderCache(c *ctx.Context) {
	req := runner.RunClearProviderCacheReq{}
	if err := c.BindJSON(&req); err != nil {
		c.Error(err, http.StatusBadRequest)
		return
	}

	count := strings.Count(req.Source, "/")
	providerCachePath := configs.Get().Runner.AbsProviderCachePath()
	if count == 2 {
		if ok, _ := runner.DeleteProviderCache(providerCachePath, req.Source, req.Version); ok { //nolint
			return
		}
	} else if count == 1 {
		if ok, _ := runner.DeleteProviderCache(filepath.Join(providerCachePath, "registry.terraform.io"), req.Source, req.Version); ok { //nolint
			return
		}
	}
}
