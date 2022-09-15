// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handler

import (
	"cloudiac/runner"
	"cloudiac/runner/api/ctx"
	"net/http"
	"os"
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
	if count == 2 {
		fullPath := filepath.Join("./var/plugin-cache", req.Source, req.Version)
		exist, err := runner.PathExists(fullPath)
		if err == nil && exist {
			err := os.RemoveAll(fullPath)
			if err != nil {
				return
			}
		} else {
			return
		}
	} else if count == 1 {
		ioPath := filepath.Join("./var/plugin-cache/registry.terraform.io", req.Source, req.Version)
		exist, err := runner.PathExists(ioPath)
		if err == nil && exist {
			err := os.RemoveAll(ioPath)
			if err != nil {
				return
			}
		} else {
			orgPath := filepath.Join("./var/plugin-cache/registry.cloudiac.org", req.Source, req.Version)
			exist, err := runner.PathExists(orgPath)
			if err == nil && exist {
				err := os.RemoveAll(orgPath)
				if err != nil {
					return
				}
			} else {
				comPath := filepath.Join("./var/plugin-cache/iac-registry.idcos.com", req.Source, req.Version)
				exist, err := runner.PathExists(comPath)
				if err == nil && exist {
					err := os.RemoveAll(comPath)
					if err != nil {
						return
					}
				}
			}
		}
	}
}
