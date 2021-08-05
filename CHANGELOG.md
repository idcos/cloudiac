------
## v0.5.1-rc2
#### Fixed
- 修复有敏感变量时执行部署报错的问题
- 修复无组织时返回 nil 导致前端报错的问题
- 修复部署日志添加评论报“任务己存在”的问题

## v0.5.1-rc1
#### Features
- 支持配置 JWT 和 AES 的密钥

#### Changes
- 升级 gorm2.0
- 修改 repos/cloud-iac 为 repos/cloudiac
- 模板的 tfvars 和 playbook 配置只在创建环境时使用，之后模板的修改不影响环境
- 调整任务队列和任务状态的轮询间隔为 1s
- 环境非活跃时设置 ttl 不同步设置 autoDestroyAt

#### Fixed
- 修复 local VCS 分支中带 "/" 时无法正常处理的问题

#### 升级步骤
1. 升级完成后执行以下 sql，更新模板的 repo_id 和 repo_addr        
**备份数据**
```sql
update iac_template SET repo_id = replace(repo_id,'/cloud-iac/','/cloudiac/') where repo_id like '/cloud-iac/%';
update iac_template SET repo_addr = replace(repo_addr,'/repos/cloud-iac/','/repos/cloudiac/') where repo_addr like '%/repos/cloud-iac/%';
```

2. 删除 deleted_at 字段
gorm 统一使用了 deleted_at_t 字段进行软删除标识

**备份数据**
```sql
ALTER TABLE iac_env DROP COLUMN deleted_at;
ALTER TABLE iac_task DROP COLUMN deleted_at;
ALTER TABLE iac_user DROP COLUMN deleted_at;
ALTER TABLE iac_project DROP COLUMN deleted_at;
ALTER TABLE iac_template DROP COLUMN deleted_at;
```

3. 添加 SECRET_KEY 环境变量配置
`.env` 文件中添加以下内容(若己存在则不需要配置)
```
SECRET_KEY=xxxx	# 变量值根据环境进行设置
```


------
## v0.5.0 20210728
全新 0.5.0 版本发布

