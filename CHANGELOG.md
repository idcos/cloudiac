
------
## vx.x.x WIP
#### Changes
- 修改 repos/cloud-iac 为 repos/cloudiac

#### 升级流程
1. 升级完成后执行以下 sql，更新模板的 repo_id 和 repo_addr        

**备份数据**
```sql
update iac_template SET repo_id = replace(repo_id,'/cloud-iac/','/cloudiac/') where repo_id like '/cloud-iac/%';
update iac_template SET repo_addr = replace(repo_addr,'/repos/cloud-iac/','/repos/cloudiac/') where repo_addr like '%/repos/cloud-iac/%';
```

2. 修改 deleted_at 字段类型，删除 deleted_at_t 字段

**备份数据**
```sql
ALTER TABLE iac_env DROP COLUMN deleted_at;
ALTER TABLE iac_task DROP COLUMN deleted_at;
ALTER TABLE iac_user DROP COLUMN deleted_at;
ALTER TABLE iac_project DROP COLUMN deleted_at;
ALTER TABLE iac_template DROP COLUMN deleted_at;
```

------
## v0.5.0 20210728
全新 0.5.0 版本发布

