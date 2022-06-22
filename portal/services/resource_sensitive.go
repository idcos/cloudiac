// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/utils"

	"github.com/tidwall/gjson"
)

func GetSensitiveKeysFromTfPlan(content []byte) map[string][]string {
	sensitiveKeys := make(map[string][]string)

	resouces := gjson.GetBytes(content, "configuration.root_module.resources")
	variables := gjson.GetBytes(content, "configuration.root_module.variables")

	// 查找 sensitive 变量
	sensitiveNames := findSensitiveNames(variables)
	if len(sensitiveNames) == 0 {
		return sensitiveKeys
	}

	sensitiveKeys = findSensitiveStateKeys(resouces, sensitiveNames)
	return sensitiveKeys
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
		if !ok {
			sensitiveAttrs[k] = v
			continue
		}

		arrAttrs := make([]map[string]interface{}, 0)
		for _, val := range vals {
			if valMap, ok := val.(map[string]interface{}); ok {
				arrAttrs = append(arrAttrs, SensitiveAttrs(valMap, sensitiveKeys, key))
			}
		}
		sensitiveAttrs[k] = arrAttrs
	}

	return sensitiveAttrs
}

// findSensitiveNames 找出 tf 文件中定义的敏感变量
func findSensitiveNames(variables gjson.Result) []string {
	namesInTf := make([]string, 0)
	if !variables.Exists() {
		return namesInTf
	}

	for k, v := range variables.Map() {
		sensitive := v.Get("sensitive")
		if sensitive.Exists() && sensitive.Bool() {
			namesInTf = append(namesInTf, "var."+k)
		}
	}

	return namesInTf
}

// findSensitiveStateKeys 找出 state 文件中对应的 sensitive key
func findSensitiveStateKeys(resources gjson.Result, sensitiveNames []string) map[string][]string {
	sensitiveKeys := make(map[string][]string)
	if !resources.Exists() {
		return sensitiveKeys
	}

	for _, resource := range resources.Array() {
		keysInState := make([]string, 0)

		if !resource.Get("expressions").Exists() {
			continue
		}

		if !resource.Get("address").Exists() {
			continue
		}

		for k, v := range resource.Get("expressions").Map() {
			if v.IsObject() {
				if findSensitiveStateKeyInObjRes(v, sensitiveNames) {
					keysInState = append(keysInState, k)
				}
			}

			if v.IsArray() {
				keys := findSensitiveStateKeysInArrRes(v.Array(), k, sensitiveNames)

				keysInState = append(keysInState, keys...)
			}
		}

		if len(keysInState) > 0 {
			sensitiveKeys[resource.Get("address").String()] = keysInState
		}
	}

	return sensitiveKeys
}

func findSensitiveStateKeyInObjRes(obj gjson.Result, sensitiveNames []string) bool {
	references := obj.Get("references")
	if !references.Exists() {
		return false
	}

	for _, reference := range references.Array() {
		if utils.InArrayStr(sensitiveNames, reference.String()) {
			return true
		}
	}

	return false
}

func findSensitiveStateKeysInArrRes(arr []gjson.Result, key string, sensitiveNames []string) []string {
	keysInState := make([]string, 0)

	for _, res := range arr {
		for k, v := range res.Map() {
			if findSensitiveStateKeyInObjRes(v, sensitiveNames) {
				keysInState = append(keysInState, key+"->"+k)
			}
		}
	}

	return keysInState
}
