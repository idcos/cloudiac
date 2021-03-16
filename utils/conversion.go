package utils

import (
	"fmt"
	"strconv"
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
