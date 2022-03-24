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

func TestGetResShowName(t *testing.T) {
	addr := "address"
	const testIp string = "localhost"
	tests := []struct {
		name string
		want string
		args map[string]interface{}
	}{
		{
			name: "get public ip",
			want: testIp,
			args: map[string]interface{}{
				"public_ip": testIp,
				"name":      "cloud_j",
				"tags":      []string{"TerraformTest-disk", "TerraformTest-cpu"},
				"id":        "i-bp14yix4r13x4fg",
			},
		},
		{
			name: "get name without tags",
			want: "cloud_j",
			args: map[string]interface{}{
				//"public_ip": testIp,
				"name": "cloud_j",
				//"tags": []string{"TerraformTest-disk", "TerraformTest-cpu"},
				"id": "i-bp14yix4r13x4fg",
			},
		},
		{
			name: "get empty name",
			want: "address(i-bp14yix4r13x4fg)",
			args: map[string]interface{}{
				//"public_ip": testIp,
				"name": "",
				"tags": []string{"TerraformTest-disk", "TerraformTest-cpu"},
				"id":   "i-bp14yix4r13x4fg",
			},
		},
		{
			name: "get name with tags in array type",
			want: "cloud_j(TerraformTest-disk,TerraformTest-cpu)",
			args: map[string]interface{}{
				//"public_ip": testIp,
				"name": "cloud_j",
				"tags": []string{"TerraformTest-disk", "TerraformTest-cpu"},
				"id":   "i-bp14yix4r13x4fg",
			},
		},
		{
			name: "get name with tags in map type",
			want: "cloud_j(name=TerraformTest-disk,type=ssd)",
			args: map[string]interface{}{
				//"public_ip": testIp,
				"name": "cloud_j",
				"tags": map[string]string{"name": "TerraformTest-disk", "type": "ssd"},
				"id":   "i-bp14yix4r13x4fg",
			},
		},
		{
			name: "get name with tags in empty map type",
			want: "cloud_j",
			args: map[string]interface{}{
				"name": "cloud_j",
				//"public_ip": testIp,
				"tags": make(map[string]string),
				"id":   "i-bp14yix4r13x4fg",
			},
		},
		{
			name: "get name with tags in null type",
			want: "cloud_j",
			args: map[string]interface{}{
				"name": "cloud_j",
				//"public_ip": testIp,
				"tags": nil,
				"id":   "i-bp14yix4r13x4fg",
			},
		},
		{
			name: "name out of the rules",
			want: "address(i-bp14yix4r13x4fg)",
			args: map[string]interface{}{
				//"name":      "cloud_j",
				//"public_ip": testIp,
				"tags": []string{"TerraformTest-disk", "TerraformTest-cpu"},
				"id":   "i-bp14yix4r13x4fg",
			},
		},
		{
			name: "name out of the rules without id",
			want: "address",
			args: map[string]interface{}{
				//"name":      "cloud_j",
				//"public_ip": testIp,
				"tags": []string{"TerraformTest-disk", "TerraformTest-cpu"},
				//"id":        "i-bp14yix4r13x4fg",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			showName := GetResShowName(tt.args, addr)
			if tt.want != showName {
				t.Errorf("test case %s error, want in expect: %s, want in reality: %s\n",
					tt.name,
					tt.want,
					showName)
			}
		})
	}
}
