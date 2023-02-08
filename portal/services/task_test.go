// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/policy"
	"cloudiac/portal/models/resps"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testStateJson = `
{
    "format_version": "0.1",
    "terraform_version": "0.14.11",
    "values": {
        "outputs": {
            "public_ip": {
                "sensitive": false,
                "value": [
                    "1.2.3.4"
                ]
            }
        },
        "root_module": {
            "resources": [
                {
                    "address": "alicloud_instance.web[key]",
                    "mode": "managed",
                    "type": "alicloud_instance",
                    "name": "web",
                    "index": "key",
                    "provider_name": "registry.terraform.io/aliyun/alicloud",
                    "schema_version": 0
				},
                {
                    "address": "alicloud_instance.web[0]",
                    "mode": "managed",
                    "type": "alicloud_instance",
                    "name": "web",
                    "index": 0,
                    "provider_name": "registry.terraform.io/aliyun/alicloud",
                    "schema_version": 0,
                    "values": {
                        "allocate_public_ip": null,
                        "auto_release_time": "",
                        "auto_renew_period": null,
                        "availability_zone": "cn-beijing-c",
                        "credit_specification": "Standard",
                        "data_disks": [],
                        "deletion_protection": false,
                        "description": "",
                        "dry_run": false,
                        "force_delete": null,
                        "host_name": "iZ2zegql1snhhgun1eh9d0Z",
                        "id": "i-2zegql1snhhgun1eh9d0",
                        "image_id": "ubuntu_18_04_64_20G_alibase_20190624.vhd",
                        "include_data_disks": null,
                        "instance_charge_type": "PostPaid",
                        "instance_name": "cloudiac-example-qa",
                        "instance_type": "ecs.t5-lc1m1.small",
                        "internet_charge_type": "PayByTraffic",
                        "internet_max_bandwidth_in": 200,
                        "internet_max_bandwidth_out": 1,
                        "io_optimized": null,
                        "is_outdated": null,
                        "key_name": "",
                        "kms_encrypted_password": null,
                        "kms_encryption_context": null,
                        "password": "",
                        "period": null,
                        "period_unit": null,
                        "private_ip": "192.168.0.217",
                        "public_ip": "47.93.102.178",
                        "renewal_status": null,
                        "resource_group_id": "",
                        "role_name": "",
                        "security_enhancement_strategy": null,
                        "security_groups": [
                            "sg-2ze0kub9scdir50230yk"
                        ],
                        "spot_price_limit": 0,
                        "spot_strategy": "NoSpot",
                        "status": "Running",
                        "subnet_id": "vsw-2zekn3er4lwdn2acoqu2z",
                        "system_disk_auto_snapshot_policy_id": "",
                        "system_disk_category": "cloud_efficiency",
                        "system_disk_description": null,
                        "system_disk_name": null,
                        "system_disk_performance_level": "",
                        "system_disk_size": 40,
                        "tags": {},
                        "timeouts": null,
                        "user_data": "Content-Type: multipart/mixed; boundary=\"MIMEBOUNDARY\"\nMIME-Version: 1.0\r\n\r\n--MIMEBOUNDARY\r\nContent-Disposition: attachment; filename=\"_cloudiac_cloud_init.sh\"\r\nContent-Transfer-Encoding: 7bit\r\nContent-Type: text/x-shellscript\r\nMime-Version: 1.0\r\n\r\n#!/bin/sh\nmkdir -p /root/.ssh/ && \\\necho 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCsJsaxzT+J3OQTbap46sQvtztLa7Lu1+tc9WlnEH5gR2WdR/DC/rVqrPqIq6KyTlOSbSg6MSLaraKy/eJ0tbls1j9gt8HnIA9soajJXeMV2sgu4DPsIIoRvIvSy1Jg2puODB+MCitz5HPuLS57eyZhCNpSNaUl0QFmQOB0m7Xp1Qe2n8ZeuM2/CMCfBn0V6EdICsC7YpAdwlwJMM6vXfHI7EhrUPEXrczNoT96KCCtksCJR2zEK9mvZm4H1S6uhOo1+MxoApiuM+cg5bv8JVfKrxRmcvm8f4+6VeD1BxQWkLXanUBtJN3V4k59mfRj+K26Tl8TJmyKYyMpCOwGx5md61WKDVu67rzfIul1aTnessHfv5SgGqeIRNZZjsJ2gaX8zgCUcHEt/4ppWf0D9rY8rmetDcvAuY8wA33vSS7M6vX0iQJQscipJb+DHU07Vl/Cxmwnap5/ObzO8CW9E2xN/V+ueCwLv3+E1S7sOaEkrSrUFbafzgfK5+JFOMHRmH4352o8Z6ABF1GfuUvvKW8T5r5p144+Qa0E0173dPnObpsEOSsXUqtrhmLsYRRbVJ1ddPGkkUnAmzP7L1TIDzdWxTw07l7z1H7blNkX9W0uINKpfTJkgN9PS+IJC6oTbWxla/UviKaj/uAZJHMkokHl/LOVbZ2WJ7EBisLhoonAGQ== CloudIaC' >> /root/.ssh/authorized_keys && \\\nchmod 0600 /root/.ssh/authorized_keys\n\n\r\n--MIMEBOUNDARY--\r\n",
                        "volume_tags": {},
                        "vswitch_id": "vsw-2zekn3er4lwdn2acoqu2z"
                    },
                    "depends_on": [
                        "alicloud_security_group.default",
                        "alicloud_vpc.default",
                        "alicloud_vswitch.default",
                        "data.cloudinit_config.cloudiac"
                    ]
                },
                {
                    "address": "data.cloudinit_config.cloudiac",
                    "mode": "data",
                    "type": "cloudinit_config",
                    "name": "cloudiac",
                    "provider_name": "registry.terraform.io/hashicorp/cloudinit",
                    "schema_version": 0,
                    "values": {
                        "base64_encode": false,
                        "boundary": "MIMEBOUNDARY",
                        "gzip": false,
                        "id": "1686996554",
                        "part": [
                            {
                                "content": "",
                                "content_type": "text/x-shellscript",
                                "filename": "_cloudiac_cloud_init.sh",
                                "merge_type": ""
                            }
                        ],
                        "rendered": ""
                    }
                }
            ]
        }
    }
}
`

func TestParseStateJson(t *testing.T) {
	state, err := UnmarshalStateJson([]byte(testStateJson))
	if err != nil {
		t.Fatal(err)
	}

	outputs := state.Values.Outputs
	variable := outputs["public_ip"]
	if !assert.Equal(t, []interface{}{"1.2.3.4"}, variable.Value) {
		t.FailNow()
	}

	res := state.Values.RootModule.Resources
	if !assert.Equal(t, res[0].Address, "alicloud_instance.web[key]") {
		t.FailNow()
	}
	if !assert.Equal(t, res[1].Address, "alicloud_instance.web[0]") {
		t.FailNow()
	}
}

var tfconfigJson = `
{
  "alicloud_instance": [
    {
      "id": "alicloud_instance.instance",
      "name": "instance",
      "module_name": "root",
      "source": "main.tf",
      "plan_root": "./",
      "line": 50,
      "type": "alicloud_instance",
      "config": {
        "availability_zone": "cn-beijing-a",
        "count": 1,
        "image_id": "ubuntu_18_04_64_20G_alibase_20190624.vhd",
        "instance_name": "tf_jack_instance",
        "instance_type": "ecs.t5-lc1m1.small",
        "internet_max_bandwidth_out": 0,
        "password": "Hello123",
        "security_groups": "${alicloud_security_group.jack_secg.*.id}",
        "system_disk_category": "cloud_efficiency",
        "vswitch_id": "${alicloud_vswitch.jack_vsw.id}"
      },
      "skip_rules": null,
      "max_severity": "",
      "min_severity": ""
    }
  ]
}
`

func TestUnmarshalTfParseJson(t *testing.T) {
	type args struct {
		bs []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *resps.TfParse
		wantErr bool
	}{
		{
			name: "Test parse terrascan config",
			args: args{bs: []byte(tfconfigJson)},
			want: &resps.TfParse{
				"alicloud_instance": resps.TSResources{
					{
						Id:         "alicloud_instance.instance",
						Name:       "instance",
						ModuleName: "root",
						Source:     "main.tf",
						PlanRoot:   "./",
						Line:       50,
						Type:       "alicloud_instance",
						Config: map[string]interface{}{
							"availability_zone":          "cn-beijing-a",
							"count":                      1,
							"image_id":                   "ubuntu_18_04_64_20G_alibase_20190624.vhd",
							"instance_name":              "tf_jack_instance",
							"instance_type":              "ecs.t5-lc1m1.small",
							"internet_max_bandwidth_out": 0,
							"password":                   "Hello123",
							"security_groups":            "${alicloud_security_group.jack_secg.*.id}",
							"system_disk_category":       "cloud_efficiency",
							"vswitch_id":                 "${alicloud_vswitch.jack_vsw.id}",
						},
						SkipRules:   nil,
						MaxSeverity: "",
						MinSeverity: "",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalTfParseJson(tt.args.bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalTfParseJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotJson, err := json.Marshal(got)
			wantJson, err2 := json.Marshal(tt.want)
			if err != nil || err2 != nil || string(gotJson) != string(wantJson) {
				t.Errorf("UnmarshalTfParseJson() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

var tsresultJson = `
{
  "results": {
    "violations": [
      {
        "rule_name": "cloudfrontNoGeoRestriction",
        "description": "Ensure that geo restriction is enabled for your Amazon ...",
        "rule_id": "AWS.CloudFront.Network Security.Low.0568",
        "severity": "LOW",
        "category": "Network Security",
        "resource_name": "s3-distribution-TLS-v1",
        "resource_type": "aws_cloudfront_distribution",
        "file": "terrascan-492583054.tf",
        "line": 7
      }
    ],
    "scan_summary": {
      "low": 0,
      "medium": 0,
      "high": 0
    }
  }
}
`

func TestUnmarshalTfResultJson(t *testing.T) {
	type args struct {
		bs []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *policy.TsResultJson
		wantErr bool
	}{
		{
			name: "Test parse terrascan result",
			args: args{bs: []byte(tsresultJson)},
			want: &policy.TsResultJson{
				Results: policy.TsResult{
					ScanSummary: policy.ScanSummary{
						Low:    0,
						Medium: 0,
						High:   0,
					},
					Violations: []policy.Violation{
						{
							RuleName:     "cloudfrontNoGeoRestriction",
							Description:  "Ensure that geo restriction is enabled for your Amazon ...",
							RuleId:       "AWS.CloudFront.Network Security.Low.0568",
							Severity:     "LOW",
							Category:     "Network Security",
							ResourceName: "s3-distribution-TLS-v1",
							ResourceType: "aws_cloudfront_distribution",
							File:         "terrascan-492583054.tf",
							Line:         7,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := policy.UnmarshalTfResultJson(tt.args.bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalTfResultJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotJson, err := json.Marshal(got)
			wantJson, err2 := json.Marshal(tt.want)
			if err != nil || err2 != nil || string(gotJson) != string(wantJson) {
				t.Errorf("UnmarshalTfResultJson() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
