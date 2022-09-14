package handler

import (
	"cloudiac/runner"
	"cloudiac/runner/api/ctx"
	"fmt"
	"net/http"
	"os"
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
		fullPath := fmt.Sprintf("/usr/yunji/cloudiac/var/plugin-cache/%s/%s", req.Source, req.Version)
		exist, err := runner.PathExists(fullPath)
		if err == nil && exist == true {
			err := os.RemoveAll(fullPath)
			if err != nil {
				return
			}
		} else {
			return
		}
	} else if count == 1 {
		ioPath := fmt.Sprintf("/usr/yunji/cloudiac/var/plugin-cache/registry.terraform.io/%s/%s", req.Source, req.Version)
		exist, err := runner.PathExists(ioPath)
		if err == nil && exist == true {
			err := os.RemoveAll(ioPath)
			if err != nil {
				return
			}
		} else {
			orgPath := fmt.Sprintf("/usr/yunji/cloudiac/var/plugin-cache/registry.cloudiac.org/%s/%s", req.Source, req.Version)
			exist, err := runner.PathExists(orgPath)
			if err == nil && exist == true {
				err := os.RemoveAll(orgPath)
				if err != nil {
					return
				}
			} else {
				comPath := fmt.Sprintf("/usr/yunji/cloudiac/var/plugin-cache/iac-registry.idcos.com/%s/%s", req.Source, req.Version)
				exist, err := runner.PathExists(comPath)
				if err == nil && exist == true {
					err := os.RemoveAll(comPath)
					if err != nil {
						return
					}
				}
			}
		}
	} else {
		return
	}

}
