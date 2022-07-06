// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/utils"
	"testing"
)

const tfplanCase01 = `
{
	"format_version": "0.1",
	"terraform_version": "0.15.5",
	"configuration": {
	  "provider_config": {
		"module.m.module.m2:docker": {
		  "name": "docker",
		  "version_constraint": "~> 2.16.0",
		  "module_address": "module.m.module.m2"
		}
	  },
	  "root_module": {
		"module_calls": {
		  "m": {
			"source": "./m",
			"expressions": {
			  "m_container_name": {
				"references": [
				  "var.root_container_name"
				]
			  }
			},
			"module": {
			  "module_calls": {
				"m2": {
				  "source": "./m2",
				  "expressions": {
					"container_name": {
					  "references": [
						"var.m_container_name"
					  ]
					},
					"my_port": {
					  "references": [
						"var.m2_my_port"
					  ]
					}
				  },
				  "module": {
					"outputs": {
					  "container_id": {
						"sensitive": true,
						"expression": {
						  "references": [
							"docker_container.nginx"
						  ]
						},
						"description": "ID of the Docker container"
					  },
					  "image_id": {
						"expression": {
						  "references": [
							"docker_image.nginx"
						  ]
						},
						"description": "ID of the Docker image"
					  }
					},
					"resources": [
					  {
						"address": "docker_container.nginx",
						"mode": "managed",
						"type": "docker_container",
						"name": "nginx",
						"provider_config_key": "m2:docker",
						"expressions": {
						  "image": {
							"references": [
							  "docker_image.nginx"
							]
						  },
						  "name": {
							"references": [
							  "var.container_name"
							]
						  },
						  "ports": [
							{
							  "external": {
								"references": [
								  "var.my_port"
								]
							  },
							  "internal": {
								"constant_value": 80
							  }
							}
						  ]
						},
						"schema_version": 2
					  },
					  {
						"address": "docker_image.nginx",
						"mode": "managed",
						"type": "docker_image",
						"name": "nginx",
						"provider_config_key": "m2:docker",
						"expressions": {
						  "keep_locally": {
							"constant_value": false
						  },
						  "name": {
							"constant_value": "nginx:latest"
						  }
						},
						"schema_version": 0
					  }
					],
					"variables": {
					  "container_name": {
						"default": "example-docker",
						"description": "Value of the name for the Docker container"
					  },
					  "my_port": {
						"default": 8001,
						"description": "my port"
					  }
					}
				  }
				}
			  },
			  "variables": {
				"m2_my_port": {
				  "default": 8000,
				  "sensitive": true
				},
				"m_container_name": {
				  "default": "example-m"
				}
			  }
			}
		  }
		},
		"variables": {
		  "root_container_name": {
			"default": "example-abcd",
			"sensitive": true
		  }
		}
	  }
	}
  }`

func TestGetSensitiveKeysFromTfPlanCase01(t *testing.T) {
	result := GetSensitiveKeysFromTfPlan([]byte(tfplanCase01))

	if result == nil {
		t.Error("解析失败")
	}

	if _, ok := result["module.m.module.m2.docker_container.nginx"]; !ok {
		t.Error("敏感变量查找失败")
	}

	v := result["module.m.module.m2.docker_container.nginx"]
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
		"address": "module.m.module.m2.docker_container.nginx",
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
		"address":  "module.m.module.m2.docker_container.nginx",
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
	"format_version": "1.1",
	"terraform_version": "1.2.2",
	"variables": {
	  "ami_id": {
		"value": "ami_iiid"
	  },
	  "filename": {
		"value": "aaaaa"
	  }
	},
	"configuration": {
	  "provider_config": {
		"docker": {
		  "name": "docker",
		  "full_name": "registry.terraform.io/hashicorp/random",
		  "version_constraint": "~> 3.3.2"
		},
		"module.none:local": {
		  "name": "local",
		  "full_name": "registry.terraform.io/hashicorp/local",
		  "version_constraint": "2.2.3",
		  "module_address": "module.none"
		},
		"random": {
		  "name": "random",
		  "full_name": "registry.terraform.io/hashicorp/random"
		}
	  },
	  "root_module": {
		"resources": [
		  {
			"address": "random_id.server",
			"mode": "managed",
			"type": "random_id",
			"name": "server",
			"provider_config_key": "random",
			"expressions": {
			  "byte_length": {
				"constant_value": 8
			  },
			  "keepers": {
				"references": ["var.ami_id"]
			  }
			},
			"schema_version": 0
		  }
		],
		"module_calls": {
		  "none": {
			"source": "./m",
			"expressions": {
			  "filename": {
				"references": ["var.filename"]
			  }
			},
			"module": {
			  "resources": [
				{
				  "address": "local_file.literature",
				  "mode": "managed",
				  "type": "local_file",
				  "name": "literature",
				  "provider_config_key": "module.none:local",
				  "expressions": {
					"content": {
					  "constant_value": "Sun Tzu said: The art of war is of vital importance to the State.\nIt is a matter of life and death, a road either to safety or to\nruin. Hence it is a subject of inquiry which can on no account be\nneglected....\n"
					},
					"filename": {
					  "references": ["var.filename"]
					}
				  },
				  "schema_version": 0
				}
			  ],
			  "variables": {
				"filename": {
				  "default": "bbbb"
				}
			  }
			}
		  }
		},
		"variables": {
		  "ami_id": {
			"default": "ami_iiid",
			"description": "Value of the name for the Docker container",
			"sensitive": true
		  },
		  "filename": {
			"default": "aaaaa",
			"sensitive": true
		  }
		}
	  }
	}
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
		"keepers":  "keepers_value",
		"address":  "random_id.server",
	}

	sKeys := GetSensitiveKeysFromTfPlan([]byte(tfplanCase02))
	addr := attrs["address"].(string)
	result := SensitiveAttrs(attrs, sKeys[addr], "")
	if result["filename"] != "name_value" {
		t.Errorf("filename attr error: %v\n", result["filename"])
	}

	if result["keepers"] != "(sensitive value)" {
		t.Errorf("keepers attr error: %v\n", result["keepers"])
	}
}
