
------
## v0.5.1 WIP
#### Changes
- 修改 repos/cloud-iac 为 repos/cloudiac

#### 升级流程
1. 升级完成后执行以下 sql，修改模板的 repo_id
```sql
update iac_template SET repo_id = replace(repo_id, '/cloud-iac/', '/cloudiac/') where repo_id like '/cloud-iac/%';
```

------
## v0.5.0 20210728
全新 0.5.0 版本发布

