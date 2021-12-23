package accurics

instanceWithNoVpc[retVal] {
	instance := input.alicloud_instance[_]
	not instance.config.vswitch_id
    rc = "ewogICJhd3NfdnBjIjogewogICAgImFjY3VyaWNzX3ZwYyI6IHsKICAgICAgImNpZHJfYmxvY2siOiAiPGNpZHJfYmxvY2s+IiwKICAgICAgImVuYWJsZV9kbnNfc3VwcG9ydCI6ICI8ZW5hYmxlX2Ruc19zdXBwb3J0PiIsCiAgICAgICJlbmFibGVfZG5zX2hvc3RuYW1lcyI6ICI8ZW5hYmxlX2Ruc19ob3N0bmFtZXM+IgogICAgfQogIH0KfQ=="
    traverse = ""
    message = "建议您创建一个专有网络，选择自有 IP 地址范围、划分网段、配置路由表和网关等。然后将重要的数据存储在一个跟互联网网络完全隔离的内网环境，日常可以用弹性IP（EIP）或者跳板机的方式对数据进行管理。具体步骤请参见创建专有网络。"
    retVal := {"Id": instance.id, "ReplaceType": "add", "CodeType": "resource", "Traverse": traverse, "Attribute": "", "AttributeDataType": "resource", "Expected": rc, "Actual": null, "Message": message}
}
