// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/utils"
	"fmt"

	"github.com/tidwall/gjson"
)

func GetSensitiveKeysFromTfPlan(content []byte) map[string][]string {
	sensitiveKeys := make(map[string][]string)

	rootModule := gjson.GetBytes(content, "configuration.root_module")
	// 查找 sensitive 变量
	sensitiveNames := findSensitiveNames(rootModule)
	if len(sensitiveNames) == 0 {
		return sensitiveKeys
	}

	sensitiveKeys = findSensitiveStateKeys(rootModule, sensitiveNames)
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

func getSensitiveVars(vars gjson.Result) []string {
	sNames := make([]string, 0)

	if !vars.Exists() {
		return sNames
	}

	for k, v := range vars.Map() {
		sensitive := v.Get("sensitive")
		if sensitive.Exists() && sensitive.Bool() {
			sNames = append(sNames, "var."+k)
		}
	}

	return sNames
}

// findSensitiveNames 找出 tf 文件中定义的敏感变量
func findSensitiveNames(rootModule gjson.Result) []string {
	rootVars := rootModule.Get("variables")

	// root module variables
	sNames := getSensitiveVars(rootVars)

	// child module variables
	children := rootModule.Get("module_calls")
	sNames = findSensitiveNamesInChildren(sNames, children)

	return sNames
}

func updateSensitiveNamesByExpressions(sNames []string, expressions gjson.Result) []string {
	if !expressions.Exists() {
		return sNames
	}

	for k, v := range expressions.Map() {
		if findSensitiveStateKeyInObjRes(v, sNames) {
			sNames = append(sNames, "var."+k)
		}
	}

	return sNames
}

func findSensitiveNamesInChildren(sNames []string, children gjson.Result) []string {
	if !children.Exists() {
		return sNames
	}

	for _, child := range children.Map() {
		vars := child.Get("module.variables")
		sNames = append(sNames, getSensitiveVars(vars)...)

		// update sensitive names: 当前模块没有设置sensitive，但是上层模块设置了 sensitive的变量
		sNames = updateSensitiveNamesByExpressions(sNames, child.Get("expressions"))

		sNames = findSensitiveNamesInChildren(sNames, child.Get("module.module_calls"))
	}

	return sNames
}

func getSensitiveStateKeysInResource(resources gjson.Result, sensitiveNames []string, sKeys map[string][]string, modulePrefix string) {
	if !resources.Exists() {
		return
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
			k := resource.Get("address").String()
			// 补充资源 address 的 module 前缀
			if modulePrefix != "" {
				k = fmt.Sprintf("%s.%s", modulePrefix, k)
			}

			sKeys[k] = keysInState
		}
	}
}

// findSensitiveStateKeys 找出 state 文件中对应的 sensitive key
func findSensitiveStateKeys(rootModule gjson.Result, sensitiveNames []string) map[string][]string {
	sensitiveKeys := make(map[string][]string)
	// root module resources
	resources := rootModule.Get("resources")
	getSensitiveStateKeysInResource(resources, sensitiveNames, sensitiveKeys, "")

	// children module resources
	children := rootModule.Get("module_calls")
	findSensitiveStateKeysInChildren(children, sensitiveNames, sensitiveKeys, "")

	return sensitiveKeys
}

func findSensitiveStateKeysInChildren(children gjson.Result, sensitiveNames []string, sKeys map[string][]string, modulePrefix string) {
	if !children.Exists() {
		return
	}

	for mName, child := range children.Map() {
		mName = fmt.Sprintf("module.%s", mName)
		if modulePrefix != "" {
			mName = fmt.Sprintf("%s.%s", modulePrefix, mName)
		}

		resources := child.Get("module.resources")
		getSensitiveStateKeysInResource(resources, sensitiveNames, sKeys, mName)

		findSensitiveStateKeysInChildren(child.Get("module.module_calls"), sensitiveNames, sKeys, mName)
	}
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
