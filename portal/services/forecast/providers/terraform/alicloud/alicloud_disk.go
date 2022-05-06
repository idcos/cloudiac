// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

//func getDiskRegistryItem() *schema.RegistryItem {
//	return &schema.RegistryItem{
//		Name:  "alicloud_disk",
//		Notes: []string{},
//		RFunc: NewDisk,
//	}
//}
//
//func NewDisk(d *schema.ResourceData) *schema.Resource {
//	region := d.Get("region").String()
//
//	a := &alicloud.Disk{
//		Address:  d.Address,
//		Provider: d.ProviderName,
//		Region:   region,
//	}
//
//	return a.BuildResource()
//
//}
