// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package policy

import (
	"cloudiac/portal/consts/e"
	"os"
	"testing"
)

func Test_getGitUrl1(t *testing.T) {
	type args struct {
		repoAddr string
		token    string
		version  string
		subDir   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"template local",
			args{"http://10.0.2.135/repos/cloudiac/terraform-alicloud-disk.git", "", "master", ""},
			"http://10.0.2.135/repos/cloudiac/terraform-alicloud-disk.git?ref=master",
		},
		{
			"template",
			args{"http://gitlab.idcos.com/iacsample/cloudiac-example.git", "the_token", "master", ""},
			"http://token:the_token@gitlab.idcos.com/iacsample/cloudiac-example.git?ref=master",
		},
		{
			"env",
			args{"http://token:the_token@gitlab.idcos.com/iacsample/cloudiac-example.git", "", "master", "ansible"},
			"http://token:the_token@gitlab.idcos.com/iacsample/cloudiac-example.git//ansible?ref=master",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGitUrl(tt.args.repoAddr, tt.args.token, tt.args.version, tt.args.subDir); got != tt.want {
				t.Errorf("getGitUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseMeta(t *testing.T) {
	type args struct {
		meta string
		rego string
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{"valid meta", args{`{
			"root":"abc",
			"category": "INFRASTRUCTURE SECURITY",
			"file": "policy.rego",
			"id": "po-c4ec520smpvc0c2qsq2g",
			"name": "instanceWithNoVpc",
			"policy_type": "alicloud",
			"reference_id": "cloudiac_alicloud_security_p001",
			"resource_type": "alicloud_instance",
			"severity": "medium",
			"version": 1
		}`, ""}, 0},
		{"invalid severity meta json", args{`{
			"category": "INFRASTRUCTURE SECURITY",
			"file": "policy.rego",
			"id": "po-c4ec520smpvc0c2qsq2g",
			"name": "instanceWithNoVpc",
			"policy_type": "alicloud",
			"reference_id": "cloudiac_alicloud_security_p001",
			"resource_type": "alicloud_instance",
			"severity": "WRONG_SEVRITY",
			"version": 1
		}`, ""}, e.PolicyMetaInvalid},
		{"valid rego", args{"", `package cloudiac
		# @id: cloudiac_alicloud_security_p001
		# @name: 策略名称A
		# @description: 这是策略的描述
		# @policy_type: alicloud
		# @resource_type: aliyun_instance
		# @severity: MEDIUM
		# @label: cat1,cat2
		# @fix_suggestion:
		# Terraform 代码去掉 associate_public_ip_address 配置
		# resource "aws_instance" "bar" {
		#  ...
		# - associate_public_ip_address = true
		# }
		# @fix_suggestion_end

		instanceWithNoVpc[instance.id] {
			instance := input.alicloud_instance[_]
			not instance.config.vswitch_id
		}

		instanceWithNoVpc[instance.id] {
			instance := input.alicloud_instance[_]
			object.get(instance.config, "vswitch_id", "undefined") == "undefined"
		}`}, 0},
	}

	defer os.Remove("meta.json")
	defer os.Remove("policy.rego")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metaFileName := tt.name + "meta.json"
			policyFileName := tt.name + "policy.rego"
			writeFile(tt.args.rego, policyFileName)
			defer os.Remove(policyFileName)
			if tt.args.meta == "" {
				if _, got := ParseMeta(policyFileName, ""); (got == nil && tt.want != 0) || (got != nil && got.Code() != tt.want) {
					t.Errorf("parseMeta() = %v, want %v", got, tt.want)
				}
			} else {
				writeFile(tt.args.meta, metaFileName)
				defer os.Remove(metaFileName)
				if _, got := ParseMeta(policyFileName, metaFileName); (got == nil && tt.want != 0) || (got != nil && got.Code() != tt.want) {
					t.Errorf("parseMeta() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func writeFile(cont string, filePath string) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, _ = f.WriteString(cont)
}
