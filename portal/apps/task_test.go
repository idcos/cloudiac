// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"reflect"
	"testing"
)

func TestGetResourcesGraphModule(t *testing.T) {
	GetResourcesGraphModule([]services.Resource{
		{
			Resource: models.Resource{
				Address: "module.tf-instances.alicloud_security_group.default",
			},
		},
		{
			Resource: models.Resource{
				Address: "module.tf-instances.alicloud_security_group_rule.allow_all_tcp",
			},
		},
		{
			Resource: models.Resource{
				Address: "module.tf-instances.alicloud_vpc.vpc",
			},
		},
		{
			Resource: models.Resource{
				Address: "module.tf-instances.alicloud_vswitch.vsw",
			},
		},
		{
			Resource: models.Resource{
				Address: "module.tf-instances.module.tf-instances.alicloud_instance.this[0]",
			},
		},
		{
			Resource: models.Resource{
				Address: "module.tf-instances.module.tf-instances.alicloud_instance.this[1]",
			},
		},
	})
	type args struct {
		rs []services.Resource
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResourcesGraphModule(tt.args.rs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResourcesGraphModule() = %v, want %v", got, tt.want)
			}
		})
	}
}
