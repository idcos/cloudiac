// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
)

func GetRegistryAddrStr(db *db.Session) string {
	cfg, err := GetSystemConfigByName(db, models.SysCfgNamRegistryAddr)
	if err == nil && cfg != nil && cfg.Value != "" {
		return strings.TrimRight(cfg.Value, "/")
	}
	return strings.TrimRight(configs.Get().RegistryAddr, "/")
}

// GetVcsRegistryAddr 获取用于读取 stack 和 policy 代码仓库的的 registry address
func GetVcsRegistryAddr(db *db.Session) string {
	addr := os.Getenv("CLOUDIAC_VCS_REGISTRY_ADDRESS")
	if addr != "" {
		return addr
	}

	return GetRegistryAddrStr(db)
}

type registryResp struct {
	Code          int             `json:"code"`
	Message       string          `json:"message"`
	MessageDetail string          `json:"messageDetail"`
	Result        json.RawMessage `json:"result"`
}

func RegistryGet(path string, data url.Values, result interface{}) (err error) {
	host := GetRegistryAddrStr(db.Get())
	if host == "" {
		return fmt.Errorf("registry address is empty")
	}
	return RegistryGetWithHost(host, path, data, result)
}

func VcsRegistryGet(path string, data url.Values, result interface{}) (err error) {
	host := GetVcsRegistryAddr(db.Get())
	if host == "" {
		return fmt.Errorf("stack registry address is empty")
	}
	return RegistryGetWithHost(host, path, data, result)
}

func RegistryGetWithHost(host string, path string, data url.Values, result interface{}) (err error) {
	logger := logs.Get().WithField("func", "RegistryGetWithAddr")

	if !strings.HasPrefix(path, "/api/v") {
		path = filepath.Join("/api/v1", path)
	}
	fullAddr := fmt.Sprintf("%s%s?%s", host, path, data.Encode())
	logger.Debugf("request %s", fullAddr)

	httpResp, err := utils.HttpClient().Get(fullAddr) //nolint:gosec
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
