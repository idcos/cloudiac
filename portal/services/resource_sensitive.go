// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/utils"

	"github.com/tidwall/gjson"
)

func GetSensitiveKeysFromTfPlan(content []byte) map[string][]string {
	rootModule := gjson.GetBytes(content, "resource_changes")
	// 查找 sensitive 变量
	return findSensitiveNames(rootModule)
}

func SensitiveAttrs(attrs map[string]interface{}, sensitiveKeys []string, parentKey string) map[string]interface{} {
	sensitiveAttrs := make(map[string]interface{})

	for k, v := range attrs {
		key := k
		if parentKey != "" {
			key = parentKey + "->" + k
		}
		if utils.InArrayStr(sensitiveKeys, key) {
			sensitiveAttrs[k] = "(sensitive value)"
			continue
		}

		vals, ok := v.([]interface{})
		if ok {
			arrAttrs := make([]map[string]interface{}, 0)
			for _, val := range vals {
				if valMap, ok := val.(map[string]interface{}); ok {
					arrAttrs = append(arrAttrs, SensitiveAttrs(valMap, sensitiveKeys, key))
				}
			}
			sensitiveAttrs[k] = arrAttrs
			continue
		}

		mVals, ok := v.(map[string]interface{})
		if ok {
			sensitiveAttrs[k] = SensitiveAttrs(mVals, sensitiveKeys, key)
			continue
		}

		sensitiveAttrs[k] = v
	}

	return sensitiveAttrs
}

func findSensitiveNames(resourceChanges gjson.Result) map[string][]string {
	mSensitiveNames := make(map[string][]string)

	for _, resource := range resourceChanges.Array() {
		sensitiveInfo := resource.Get("change.after_sensitive")
		if !sensitiveInfo.Exists() {
			return mSensitiveNames
		}
		if !sensitiveInfo.IsObject() {
			return mSensitiveNames
		}

		addr := resource.Get("address").String()
		sNames := findSensitiveNamesInResource(resource)
		if len(sNames) > 0 {
			mSensitiveNames[addr] = sNames
		}
	}
	return mSensitiveNames
}

func findSensitiveNamesInResource(resource gjson.Result) []string {
	sNames := make([]string, 0)
	sensitiveInfo := resource.Get("change.after_sensitive")
	if !sensitiveInfo.Exists() {
		return sNames
	}
	if !sensitiveInfo.IsObject() {
		return sNames
	}

	for k, v := range sensitiveInfo.Map() {
		if v.IsBool() && v.Bool() {
			sNames = append(sNames, k)
			continue
		}

		if v.IsArray() {
			for _, vv := range v.Array() {
				for vvk, vvv := range vv.Map() {
					if vvv.IsBool() && vvv.Bool() {
						sNames = append(sNames, k+"->"+vvk)
					}
				}
			}
		}

		if v.IsObject() {
			for vk, vv := range v.Map() {
				if vv.IsBool() && vv.Bool() {
					sNames = append(sNames, k+"->"+vk)
				}
			}
		}

	}

	return sNames
}
