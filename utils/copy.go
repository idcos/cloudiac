// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package utils

import "github.com/jinzhu/copier"

// DeepCopy 深拷贝对象
// 两个参数都需要传入指针
func DeepCopy(to interface{}, from interface{}) {
	err := copier.CopyWithOption(to, from, copier.Option{DeepCopy: true, IgnoreEmpty: true})
	if err != nil {
		panic(err)
	}
}
