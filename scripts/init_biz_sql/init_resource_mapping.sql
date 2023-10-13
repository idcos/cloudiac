-- 添加vmware 资源名称映射
INSERT INTO iac_resource_mapping (id, provider, type, code, express)
VALUES ( uuid_short(), 'registry.terraform.io/hashicorp/vsphere', 'vsphere_virtual_machine', 'name', 'name');
