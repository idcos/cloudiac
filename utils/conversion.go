package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func Str2float(s string) float64 {
	i, _ := strconv.ParseFloat(s, 64)
	return i
}

func Str2int(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

//保留两位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

//判断数组中是否存在该元素
func ArrayIsExists(s []uint, e uint) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

func GetSumArray(s []float64) float64 {
	sum := 0.0
	for _, i := range s {
		sum += i
	}
	return sum
}

//判断某个字符串是否以数组中的元素结尾
func ArrayIsHasSuffix(arr []string, v string) bool {
	for i := range arr {
		if strings.HasSuffix(v, arr[i]) {
			return true
		}
	}
	return false
}

//去掉url路径结尾的'/'
func GetUrl(address string) string {
	return strings.TrimSuffix(address, "/")
}
