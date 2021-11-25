package apps

import (
	"cloudiac/portal/models"
	"reflect"
	"testing"
)

func TestGetResourcesGraphModule(t *testing.T) {
	GetResourcesGraphModule([]models.Resource{
		{
			Address:"module.tf-instances.alicloud_security_group.default",
		},
		{
			Address:"module.tf-instances.alicloud_security_group_rule.allow_all_tcp",
		},
		{
			Address:"module.tf-instances.alicloud_vpc.vpc",
		},
		{
			Address:"module.tf-instances.alicloud_vswitch.vsw",
		},
		{
			Address:"module.tf-instances.module.tf-instances.alicloud_instance.this[0]",
		},
		{
			Address:"module.tf-instances.module.tf-instances.alicloud_instance.this[1]",
		},
	})
	type args struct {
		rs []models.Resource
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
