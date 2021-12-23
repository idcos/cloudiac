package cloudiac

## id 为策略在策略组中的唯一标识，由大小写英文字符、数字、"."、"_"、"-" 组成
## 建议按`组织_云商_资源名称/分类_编号`的格式进行命名
# @id: cloudiac_alicloud_security_p001

# @name: 策略名称A
# @description: 这是策略的描述

## 策略类型，如 aws, k8s, github, alicloud, ...
# @policy_type: alicloud

## 资源类型，如 aws_ami, k8s_pod, alicloud_instance, ...
# @resource_type: aliyun_instance

## 策略严重级别: 可选 HIGH/MEDIUM/LOW
# @severity: MEDIUM

## 策略标签，多个分类使用逗号分隔
# @label: cat1,cat2

## 策略修复建议（支持多行，以@fix_suggestion_end结束）
# @fix_suggestion:
# Terraform 代码去掉`associate_public_ip_address`配置
# ```
# resource "aws_instance" "bar" {
#  ...
# - associate_public_ip_address = true
# }
# ```
# @fix_suggestion_end

instanceWithNoVpc[instance.id] {
	instance := input.alicloud_instance[_]
	not instance.config.vswitch_id
}

instanceWithNoVpc[instance.id] {
	instance := input.alicloud_instance[_]
	object.get(instance.config, "vswitch_id", "undefined") == "undefined"
}
