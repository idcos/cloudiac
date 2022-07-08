// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/utils"
	"testing"
)

const tfplanCase01 = `{
	"format_version": "0.1",
	"terraform_version": "0.15.5",
	"resource_changes": [
	  {
		"address": "docker_container.nginx",
		"mode": "managed",
		"type": "docker_container",
		"name": "nginx",
		"provider_name": "registry.terraform.io/kreuzwerker/docker",
		"change": {
		  "actions": ["create"],
		  "before": null,
		  "after": {
			"attach": false,
			"capabilities": [],
			"cpu_set": null,
			"cpu_shares": null,
			"destroy_grace_seconds": null,
			"devices": [],
			"dns": null,
			"dns_opts": null,
			"dns_search": null,
			"domainname": null,
			"group_add": null,
			"host": [],
			"links": null,
			"log_opts": null,
			"logs": false,
			"max_retry_count": null,
			"memory": null,
			"memory_swap": null,
			"mounts": [],
			"must_run": true,
			"name": "example-name",
			"network_alias": null,
			"network_mode": null,
			"networks": null,
			"networks_advanced": [],
			"pid_mode": null,
			"ports": [
			  {
				"external": 8000,
				"internal": 80,
				"ip": "0.0.0.0",
				"protocol": "tcp"
			  }
			],
			"privileged": null,
			"publish_all_ports": null,
			"read_only": false,
			"remove_volumes": true,
			"restart": "no",
			"rm": false,
			"start": true,
			"stdin_open": false,
			"storage_opts": null,
			"sysctls": null,
			"tmpfs": null,
			"tty": false,
			"ulimit": [],
			"upload": [],
			"user": null,
			"userns_mode": null,
			"volumes": [],
			"working_dir": null
		  },
		  "after_unknown": {
			"bridge": true,
			"capabilities": [],
			"command": true,
			"container_logs": true,
			"devices": [],
			"entrypoint": true,
			"env": true,
			"exit_code": true,
			"gateway": true,
			"healthcheck": true,
			"host": [],
			"hostname": true,
			"id": true,
			"image": true,
			"init": true,
			"ip_address": true,
			"ip_prefix_length": true,
			"ipc_mode": true,
			"labels": true,
			"log_driver": true,
			"mounts": [],
			"network_data": true,
			"networks_advanced": [],
			"ports": [{}],
			"security_opts": true,
			"shm_size": true,
			"ulimit": [],
			"upload": [],
			"volumes": []
		  },
		  "before_sensitive": false,
		  "after_sensitive": {
			"capabilities": [],
			"command": [],
			"devices": [],
			"entrypoint": [],
			"env": [],
			"healthcheck": [],
			"host": [],
			"labels": [],
			"mounts": [],
			"name": true,
			"network_data": [],
			"networks_advanced": [],
			"ports": [
			  {
				"external": true
			  }
			],
			"security_opts": [],
			"ulimit": [],
			"upload": [],
			"volumes": []
		  }
		}
	  },
	  {
		"address": "docker_image.nginx",
		"mode": "managed",
		"type": "docker_image",
		"name": "nginx",
		"provider_name": "registry.terraform.io/kreuzwerker/docker",
		"change": {
		  "actions": ["create"],
		  "before": null,
		  "after": {
			"build": [],
			"force_remove": null,
			"keep_locally": false,
			"name": "nginx:latest",
			"pull_trigger": null,
			"pull_triggers": null
		  },
		  "after_unknown": {
			"build": [],
			"id": true,
			"latest": true,
			"output": true,
			"repo_digest": true
		  },
		  "before_sensitive": false,
		  "after_sensitive": {
			"build": []
		  }
		}
	  },
	  {
		"address": "docker_secret.nginx",
		"mode": "managed",
		"type": "docker_secret",
		"name": "nginx",
		"provider_name": "registry.terraform.io/kreuzwerker/docker",
		"change": {
		  "actions": ["create"],
		  "before": null,
		  "after": {
			"data": "dGVzdHRlc3Q=",
			"labels": [],
			"name": "nginx-swarm"
		  },
		  "after_unknown": {
			"id": true,
			"labels": []
		  },
		  "before_sensitive": false,
		  "after_sensitive": {
			"data": true,
			"labels": []
		  }
		}
	  }
	]
  }
`

func TestGetSensitiveKeysFromTfPlanCase01(t *testing.T) {
	result := GetSensitiveKeysFromTfPlan([]byte(tfplanCase01))

	if result == nil {
		t.Error("解析失败")
	}

	t.Logf("result: %v\n", result)

	if _, ok := result["docker_secret.nginx"]; !ok {
		t.Error("敏感变量查找失败")
	}

	v := result["docker_container.nginx"]
	if len(v) != 2 {
		t.Error("敏感变量数量错误")
	}

	if !utils.InArrayStr(v, "name") || !utils.InArrayStr(v, "ports->external") {
		t.Error("敏感变量key错误")
	}
}

func TestCase01SensitiveAttrs01(t *testing.T) {
	// name 是敏感变量
	var attrs = map[string]interface{}{
		"name":    "name_value",
		"image":   "image_value",
		"address": "docker_container.nginx",
	}

	sKeys := GetSensitiveKeysFromTfPlan([]byte(tfplanCase01))
	addr := attrs["address"].(string)
	result := SensitiveAttrs(attrs, sKeys[addr], "")
	if result["name"] != "(sensitive value)" {
		t.Errorf("name attr error: %v\n", result["name"])
	}

	if result["image"] != "image_value" {
		t.Errorf("image attr error: %v\n", result["image"])
	}
}

func TestCase01SensitiveAttrs02(t *testing.T) {
	// external 是敏感变量，但是有前缀 ports
	var attrs = map[string]interface{}{
		"external": "external_value",
		"ip":       "ip_value",
		"address":  "docker_container.nginx",
	}

	sKeys := GetSensitiveKeysFromTfPlan([]byte(tfplanCase01))
	addr := attrs["address"].(string)

	result := SensitiveAttrs(attrs, sKeys[addr], "ports")
	if result["external"] != "(sensitive value)" {
		t.Errorf("external attr error: %v\n", result["external"])
	}

	if result["ip"] != "ip_value" {
		t.Errorf("ip attr error: %v\n", result["ip"])
	}
}

func TestCase01SensitiveAttrs03(t *testing.T) {
	// name 和 external 是敏感变量，但是资源 address 不匹配，所以不加密
	var attrs = map[string]interface{}{
		"name":     "name_value",
		"external": "external_value",
		"address":  "address_value",
	}

	sKeys := GetSensitiveKeysFromTfPlan([]byte(tfplanCase01))
	addr := attrs["address"].(string)
	result := SensitiveAttrs(attrs, sKeys[addr], "")
	if result["name"] != "name_value" {
		t.Errorf("name attr error: %v\n", result["name"])
	}
	if result["external"] != "external_value" {
		t.Errorf("external attr error: %v\n", result["external"])
	}
}

var tfplanCase02 = `{
	"format_version": "0.1",
	"terraform_version": "0.15.5",
	"resource_changes": [
	  {
		"address": "module.none.local_file.literature",
		"module_address": "module.none",
		"mode": "managed",
		"type": "local_file",
		"name": "literature",
		"provider_name": "registry.terraform.io/hashicorp/local",
		"change": {
		  "actions": ["create"],
		  "before": null,
		  "after": {
			"content": "Sun Tzu said: The art of war is of vital importance to the State.\nIt is a matter of life and death, a road either to safety or to\nruin. Hence it is a subject of inquiry which can on no account be\nneglected....\n",
			"content_base64": null,
			"directory_permission": "0777",
			"file_permission": "0777",
			"filename": "aaaaa",
			"sensitive_content": null,
			"source": null
		  },
		  "after_unknown": {
			"id": true
		  },
		  "before_sensitive": false,
		  "after_sensitive": {
			"filename": true,
			"sensitive_content": true
		  }
		}
	  },
	  {
		"address": "random_id.server",
		"mode": "managed",
		"type": "random_id",
		"name": "server",
		"provider_name": "registry.terraform.io/hashicorp/random",
		"change": {
		  "actions": ["create"],
		  "before": null,
		  "after": {
			"byte_length": 8,
			"keepers": {
			  "ami_id": "ami_iiid",
			  "ami_id2": "ami_iddd"
			},
			"prefix": null
		  },
		  "after_unknown": {
			"b64_std": true,
			"b64_url": true,
			"dec": true,
			"hex": true,
			"id": true,
			"keepers": {}
		  },
		  "before_sensitive": false,
		  "after_sensitive": {
			"keepers": {
			  "ami_id": true
			}
		  }
		}
	  }
	]
  }
`

func TestGetSensitiveKeysFromTfPlanCase02(t *testing.T) {
	result := GetSensitiveKeysFromTfPlan([]byte(tfplanCase02))

	if result == nil {
		t.Error("解析失败")
	}

	t.Logf("result: %v\n", result)

	if _, ok := result["module.none.local_file.literature"]; !ok {
		t.Error("敏感变量查找失败")
	}

	if _, ok := result["random_id.server"]; !ok {
		t.Error("敏感变量查找失败")
	}

	if len(result) != 2 {
		t.Error("敏感变量数量错误")
	}
}

func TestCase02SensitiveAttrs01(t *testing.T) {
	// name 是敏感变量
	var attrs = map[string]interface{}{
		"filename": "name_value",
		"keepers":  "keepers_value",
		"address":  "module.none.local_file.literature",
	}

	sKeys := GetSensitiveKeysFromTfPlan([]byte(tfplanCase02))
	addr := attrs["address"].(string)
	result := SensitiveAttrs(attrs, sKeys[addr], "")
	if result["filename"] != "(sensitive value)" {
		t.Errorf("filename attr error: %v\n", result["filename"])
	}

	if result["keepers"] != "keepers_value" {
		t.Errorf("keepers attr error: %v\n", result["keepers"])
	}
}

func TestCase02SensitiveAttrs02(t *testing.T) {
	// name 是敏感变量
	var attrs = map[string]interface{}{
		"filename": "name_value",
		"ami_id":   "ami_id_value",
		"address":  "random_id.server",
	}

	sKeys := GetSensitiveKeysFromTfPlan([]byte(tfplanCase02))
	addr := attrs["address"].(string)
	result := SensitiveAttrs(attrs, sKeys[addr], "")
	if result["filename"] != "name_value" {
		t.Errorf("filename attr error: %v\n", result["filename"])
	}

	result = SensitiveAttrs(attrs, sKeys[addr], "keepers")
	if result["ami_id"] != "(sensitive value)" {
		t.Errorf("ami_id attr error: %v\n", result["ami_id"])
	}
}
