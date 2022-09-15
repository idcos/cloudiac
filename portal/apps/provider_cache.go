// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	runnerClear "cloudiac/runner"
	"cloudiac/utils"
	"net/http"
)

func ClearProviderCache(c *ctx.ServiceContext, form *forms.ClearProviderCacheForm) (interface{}, e.Error) {
	runners, err := services.RunnerSearch()
	if err != nil {
		return nil, err
	}

	for _, runner := range runners {

		requestUrl := utils.JoinURL(runner.Address, consts.RunnerClearProviderCache)
		req := runnerClear.RunClearProviderCacheReq{
			Source:  form.Source,
			Version: form.Version,
		}

		header := &http.Header{}
		header.Set("Content-Type", "application/json")
		timeout := int(consts.RunnerConnectTimeout.Seconds())
		_, err := utils.HttpService(requestUrl, "POST", header, req, timeout, timeout)
		if err != nil {
			return nil, e.New(e.RunnerError, err)
		}
	}

	return nil, nil
}
