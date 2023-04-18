// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func GetRegistryAddrStr(db *db.Session) string {
	cfg, err := GetSystemConfigByName(db, models.SysCfgNamRegistryAddr)
	if err == nil && cfg != nil && cfg.Value != "" {
		return strings.TrimRight(cfg.Value, "/")
	}
	return strings.TrimRight(configs.Get().RegistryAddr, "/")
}

func GetRegistryApiBase(db *db.Session) string {
	addr := GetRegistryAddrStr(db)
	return fmt.Sprintf("%s/api/v1", strings.TrimSuffix(addr, "/"))
}

type registryResp struct {
	Code          int             `json:"code"`
	Message       string          `json:"message"`
	MessageDetail string          `json:"messageDetail"`
	Result        json.RawMessage `json:"result"`
}

func RegistryGet(path string, data url.Values, result interface{}) (err error) {
	logger := logs.Get().WithField("func", "RegistryGet")

	host := GetRegistryAddrStr(db.Get())
	if host == "" {
		return fmt.Errorf("registry address is empty")
	}

	path = filepath.Join("/api/v1", path)
	fullAddr := fmt.Sprintf("%s%s?%s", host, path, data.Encode())
	logger.Debugf("request %s", fullAddr)

	httpResp, err := http.Get(fullAddr) //nolint:gosec
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	resp := registryResp{}
	if err := json.Unmarshal(body, &resp); err != nil {
		logs.Get().Debugf("response body: %s", body)
		return err
	}

	if resp.Code != 0 && resp.Code != 200 {
		logs.Get().Infof("registry response error: %+v", resp)
		return fmt.Errorf("registry error: %d: %s", resp.Code, resp.Message)
	}

	if err := json.Unmarshal(resp.Result, result); err != nil {
		logs.Get().Debugf("response body: %s", body)
		return err
	}
	return nil
}

func GetRegistryMirrorUrl(db *db.Session) string {
	hostAddr := GetRegistryAddrStr(db)
	if hostAddr == "" {
		return ""
	}
	return utils.JoinURL(hostAddr, consts.RegistryMirrorUri)
}
